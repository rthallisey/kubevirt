/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2019 Red Hat, Inc.
 *
 */

package save

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	kubevirtV1 "kubevirt.io/client-go/api/v1"

	"kubevirt.io/client-go/kubecli"
	"kubevirt.io/kubevirt/pkg/virtctl/templates"
)

const (
	COMMAND_SAVE    = "save"
	COMMAND_RESTORE = "restore"
	ARG_VM_SHORT    = "vm"
	ARG_VM_LONG     = "virtualmachine"
	ARG_VMI_SHORT   = "vmi"
	ARG_VMI_LONG    = "virtualmachineinstance"
)

func NewSaveCommand(clientConfig clientcmd.ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save vm|vmi (VM)|(VMI)",
		Short: "Save a virtual machine",
		Long: `Saves a virtual machine, which store  it. Machine state is kept in memory.
First argument is the resource type, possible types are (case insensitive, both singular and plural forms) virtualmachineinstance (vmi) or virtualmachine (vm).
Second argument is the name of the resource.`,
		Args:    templates.ExactArgs(COMMAND_SAVE, 2),
		Example: usage(COMMAND_SAVE),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := VirtCommand{
				command:      COMMAND_SAVE,
				clientConfig: clientConfig,
			}
			return c.Run(cmd, args)
		},
	}
	cmd.SetUsageTemplate(templates.UsageTemplate())
	return cmd
}

func NewRestoreCommand(clientConfig clientcmd.ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore vm|vmi (VM)|(VMI) <path_to_file>",
		Short: "Restore a virtual machine",
		Long: `Restores a virtual machine.
First argument is the resource type, possible types are (case insensitive, both singular and plural forms) virtualmachineinstance (vmi) or virtualmachine (vm).
Second argument is the name of the resource. Third argument is the path to a restoreFile`,
		Args:    templates.ExactArgs(COMMAND_RESTORE, 3),
		Example: usage(COMMAND_RESTORE),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := VirtCommand{
				command:      COMMAND_RESTORE,
				clientConfig: clientConfig,
			}
			return c.Run(cmd, args)
		},
	}
	cmd.SetUsageTemplate(templates.UsageTemplate())
	return cmd
}

func usage(cmd string) string {
	usage := fmt.Sprintf("  # %s a virtualmachine called 'myvm':\n", strings.Title(cmd))
	usage += fmt.Sprintf("  {{ProgramName}} %s vm myvm", cmd)
	return usage
}

type VirtCommand struct {
	clientConfig clientcmd.ClientConfig
	command      string
}

func (vc *VirtCommand) Run(cmd *cobra.Command, args []string) error {
	resourceType := strings.ToLower(args[0])
	resourceName := args[1]
	namespace, _, err := vc.clientConfig.Namespace()
	if err != nil {
		return err
	}

	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(vc.clientConfig)
	if err != nil {
		return fmt.Errorf("Cannot obtain KubeVirt client: %v", err)
	}

	switch vc.command {
	case COMMAND_SAVE:
		switch resourceType {
		case ARG_VM_LONG, ARG_VM_SHORT:
			vm, err := virtClient.VirtualMachine(namespace).Get(resourceName, &v1.GetOptions{})
			if err != nil {
				return fmt.Errorf("Error getting VirtualMachine %s: %v", resourceName, err)
			}
			vmiName := vm.Name
			err = virtClient.VirtualMachineInstance(namespace).Save(vmiName)
			if err != nil {
				if errors.IsNotFound(err) {
					runningStrategy, err := vm.RunStrategy()
					if err != nil {
						return fmt.Errorf("Error saving VirutalMachineInstance %s: %v", vmiName, err)
					}
					if runningStrategy == kubevirtV1.RunStrategyHalted {
						return fmt.Errorf("Error saving VirtualMachineInstance %s. VirtualMachine %s is not set to run", vmiName, vm.Name)
					}
					return fmt.Errorf("Error saving VirtualMachineInstance %s, it was not found", vmiName)

				}
				return fmt.Errorf("Error saving VirutalMachineInstance %s: %v", vmiName, err)
			}
			fmt.Printf("VMI %s was scheduled to %s\n", vmiName, vc.command)
		case ARG_VMI_LONG, ARG_VMI_SHORT:
			err = virtClient.VirtualMachineInstance(namespace).Save(resourceName)
			if err != nil {
				return fmt.Errorf("Error saving VirtualMachineInstance %s: %v", resourceName, err)
			}
			fmt.Printf("VMI %s was scheduled to %s\n", resourceName, vc.command)
		}
	case COMMAND_RESTORE:
		rFile := args[2]
		switch resourceType {
		case ARG_VM_LONG, ARG_VM_SHORT:
			vm, err := virtClient.VirtualMachine(namespace).Get(resourceName, &v1.GetOptions{})
			if err != nil {
				return fmt.Errorf("Error getting VirtualMachine %s: %v", resourceName, err)
			}
			vmiName := vm.Name
			err = virtClient.VirtualMachineInstance(namespace).Restore(vmiName, rFile)
			if err != nil {
				return fmt.Errorf("Error restoring VirtualMachineInstance %s: %v", vmiName, err)
			}
			fmt.Printf("VMI %s was scheduled to %s\n", vmiName, vc.command)
		case ARG_VMI_LONG, ARG_VMI_SHORT:
			err = virtClient.VirtualMachineInstance(namespace).Restore(resourceName, rFile)
			if err != nil {
				return fmt.Errorf("Error restoring VirtualMachineInstance %s: %v", resourceName, err)
			}
			fmt.Printf("VMI %s was scheduled to %s\n", resourceName, vc.command)
		}
	}
	return nil
}
