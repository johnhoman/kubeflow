package culler

import (
	"encoding/json"
	"fmt"
	corev1beta1 "github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

const (
	fmtJupyterKernel = "http://%[1]s.%[2]s.svc.cluster.local/notebook/%[2]s/%[1]s/api/kernels"
)

type kernelStatus struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LastActivity   string `json:"last_activity"`
	ExecutionState string `json:"execution_state"`
	Connections    int    `json:"connections"`
}

type jupyter interface {
	IsIdle(nb *corev1beta1.Notebook) (bool, error)
	LatestActivity(nb *corev1beta1.Notebook) (*time.Time, error)
}

type Client struct {
	client *http.Client
}

func (c *Client) kernels(nb *corev1beta1.Notebook) ([]kernelStatus, error) {
	u := fmt.Sprintf(fmtJupyterKernel, nb.Name, nb.Namespace)
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	var kernels = make([]kernelStatus, 0)
	if err := json.NewDecoder(resp.Body).Decode(&kernels); err != nil {
		return nil, err
	}
	return kernels, nil
}

func (c *Client) IsIdle(nb *corev1beta1.Notebook) (bool, error) {
	kernels, err := c.kernels(nb)
	if err != nil {
		return false, err
	}
	if len(kernels) == 0 {
		// no kernels means it's idle
		return true, nil
	}
	for _, item := range kernels {
		if item.ExecutionState != "idle" {
			return false, nil
		}
	}
	return true, nil
}

func (c *Client) LatestActivity(nb *corev1beta1.Notebook) (*time.Time, error) {
	kernels, err := c.kernels(nb)
	if err != nil {
		return nil, err
	}
	if len(kernels) == 0 {
		return nil, nil
	}

	latest, err := time.Parse(time.RFC3339, kernels[0].LastActivity)
	if err != nil {
		return nil, err
	}

	for _, item := range kernels {
		parsed, err := time.Parse(time.RFC3339, item.LastActivity)
		if err != nil {
			continue
		}
		if parsed.After(latest) {
			latest = parsed
		}
	}
	return &latest, nil
}

type alwaysIdle struct{}

func (a *alwaysIdle) IsIdle(nb *corev1beta1.Notebook) (bool, error) {
	return true, nil
}

func (a *alwaysIdle) LatestActivity(nb *corev1beta1.Notebook) (*time.Time, error) {
	now := time.Now().Add(-time.Hour)
	return &now, nil
}

var _ jupyter = &alwaysIdle{}

type NeverIdle struct{}

func (a *NeverIdle) IsIdle(nb *corev1beta1.Notebook) (bool, error) {
	return false, nil
}

func (a *NeverIdle) LatestActivity(nb *corev1beta1.Notebook) (*time.Time, error) {
	now := time.Now().Add(time.Hour)
	return &now, nil
}

var _ jupyter = &NeverIdle{}
