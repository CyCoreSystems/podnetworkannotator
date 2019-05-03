package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/CyCoreSystems/netdiscover/discover"
	"github.com/ericchiang/k8s"
	v1 "github.com/ericchiang/k8s/apis/core/v1"
)

// RefreshInterval is the maximum amount of time to wait to update the network annotations
const RefreshInterval = time.Minute

var thisNamespace string
var thisPod string

// PublicIPv4Annotation is the name of the annotation to be used when marking a Pod with its Public IPv4 address
const PublicIPv4Annotation = "netdiscover.cycore.io/public_ipv4"

// PublicIPv6Annotation is the name of the annotation to be used when marking a Pod with its Public IPv6 address
const PublicIPv6Annotation = "netdiscover.cycore.io/public_ipv6"

// PublicHostnameAnnotation is the name of the annotation to be used when marking a Pod with its Public Hostname
const PublicHostnameAnnotation = "netdiscover.cycore.io/public_hostname"

func main() {
	ctx := context.Background()

	thisPod = os.Getenv("POD_NAME")
	if thisPod == "" {
		log.Fatal("POD_NAME must be set")
	}
	thisNamespace = os.Getenv("POD_NAMESPACE")
	if thisNamespace == "" {
		log.Fatal("POD_NAMESPACE must be set")
	}

	var discoverer discover.Discoverer
	switch os.Getenv("CLOUD") {
	case "aws":
		discoverer = discover.NewAWSDiscoverer()
	case "azure":
		discoverer = discover.NewAzureDiscoverer()
	case "do":
		discoverer = discover.NewDigitalOceanDiscoverer()
	case "gcp":
		discoverer = discover.NewGCPDiscoverer()
	case "":
		discoverer = discover.NewDiscoverer()
	default:
		log.Fatal("unsupported CLOUD; leave empty for best-effort")
	}

	if err := AnnotatePod(ctx, discoverer); err != nil {
		log.Fatal("failed to annotate pod:", err)
	}

	for {
		time.Sleep(RefreshInterval)

		if err := AnnotatePod(ctx, discoverer); err != nil {
			log.Fatal("failed to annotate pod:", err)
		}
	}
}

// AnnotatePod adds an annotation to the Pod with its PrivateIP
func AnnotatePod(ctx context.Context, d discover.Discoverer) error {
	var publicv4, publicv6 string

	if ipv4, err := d.PublicIPv4(); err == nil {
		publicv4 = ipv4.String()
	}
	if ipv6, err := d.PublicIPv6(); err == nil {
		publicv6 = ipv6.String()
	}
	hostname, _ := d.Hostname()

	kc, err := k8s.NewInClusterClient()
	if err != nil {
		return err
	}

	var p = new(v1.Pod)

	if err = kc.Get(ctx, thisNamespace, thisPod, p); err != nil {
		return err
	}

	ann := p.GetMetadata().GetAnnotations()

	var changed bool
	changed = updateAnnotation(ann, PublicIPv4Annotation, publicv4, changed)
	changed = updateAnnotation(ann, PublicIPv6Annotation, publicv6, changed)
	changed = updateAnnotation(ann, PublicHostnameAnnotation, hostname, changed)

	if changed {
		return kc.Update(ctx, p)
	}
	return nil
}

func updateAnnotation(ann map[string]string, key, val string, changed bool) bool {
	if val == "" {
		return changed
	}

	cur, ok := ann[key]
	if ok && cur == val {
		return changed
	}
	changed = true
	ann[key] = val
	return changed
}
