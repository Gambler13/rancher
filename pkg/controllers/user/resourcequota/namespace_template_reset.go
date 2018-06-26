package resourcequota

import (
	"fmt"

	"github.com/rancher/types/apis/core/v1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	corev1 "k8s.io/api/core/v1"
	clientcache "k8s.io/client-go/tools/cache"
)

/*
templateResetController is responsible for resetting resource quota template from the namespace
when project resource quota gets reset
*/
type templateResetController struct {
	namespaces v1.NamespaceInterface
	nsIndexer  clientcache.Indexer
}

func (c *templateResetController) resetTemplate(key string, project *v3.Project) error {
	if project == nil || project.DeletionTimestamp != nil {
		return nil
	}
	if project.Spec.ResourceQuota != nil {
		return nil
	}
	projectID := fmt.Sprintf("%s:%s", project.Namespace, project.Name)
	namespaces, err := c.nsIndexer.ByIndex(nsByProjectIndex, projectID)
	if err != nil {
		return err
	}
	for _, n := range namespaces {
		ns := n.(*corev1.Namespace)
		templateID := getTemplateID(ns)
		if templateID == "" {
			continue
		}
		toUpdate := ns.DeepCopy()
		delete(toUpdate.Annotations, resourceQuotaTemplateIDAnnotation)
		delete(toUpdate.Annotations, resourceQuotaAppliedTemplateIDAnnotation)
		if _, err := c.namespaces.Update(toUpdate); err != nil {
			return err
		}
	}

	return nil
}
