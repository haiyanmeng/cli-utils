// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package apply

import (
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
	"sigs.k8s.io/cli-utils/pkg/apply/event"
)

// KubectlPrinterAdapter is a workaround for capturing progress from
// ApplyOptions. ApplyOptions were originally meant to print progress
// directly using a configurable printer. The KubectlPrinterAdapter
// plugs into ApplyOptions as a ToPrinter function, but instead of
// printing the info, it emits it as an event on the provided channel.
type KubectlPrinterAdapter struct {
	ch chan<- event.Event
}

// resourcePrinterImpl implements the ResourcePrinter interface. But
// instead of printing, it emits information on the provided channel.
type resourcePrinterImpl struct {
	applyOperation event.ApplyEventOperation
	ch             chan<- event.Event
}

// PrintObj takes the provided object and operation and emits
// it on the channel.
func (r *resourcePrinterImpl) PrintObj(obj runtime.Object, _ io.Writer) error {
	r.ch <- event.Event{
		Type: event.ApplyType,
		ApplyEvent: event.ApplyEvent{
			Type:      event.ApplyEventResourceUpdate,
			Operation: r.applyOperation,
			Object:    obj,
		},
	}
	return nil
}

type toPrinterFunc func(string) (printers.ResourcePrinter, error)

// toPrinterFunc returns a function of type toPrinterFunc. This
// is the type required by the ApplyOptions.
func (p *KubectlPrinterAdapter) toPrinterFunc() toPrinterFunc {
	return func(operation string) (printers.ResourcePrinter, error) {
		applyOperation, err := operationToApplyOperationConst(operation)
		return &resourcePrinterImpl{
			ch:             p.ch,
			applyOperation: applyOperation,
		}, err
	}
}

func operationToApplyOperationConst(operation string) (event.ApplyEventOperation, error) {
	switch operation {
	case "serverside-applied":
		return event.ServersideApplied, nil
	case "created":
		return event.Created, nil
	case "unchanged":
		return event.Unchanged, nil
	case "configured":
		return event.Configured, nil
	default:
		return event.ApplyEventOperation(0), fmt.Errorf("unknown operation %s", operation)
	}
}
