// Copyright 2022 The Inspektor Gadget authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trace

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kinvolk/inspektor-gadget/cmd/kubectl-gadget/utils"
	"github.com/kinvolk/inspektor-gadget/pkg/gadgets/tcpconnect/types"
	eventtypes "github.com/kinvolk/inspektor-gadget/pkg/types"

	"github.com/spf13/cobra"
)

func newTcpconnectCmd() *cobra.Command {
	columnsWidth := map[string]int{
		"node":      -16,
		"namespace": -16,
		"pod":       -16,
		"container": -16,
		"pid":       -7,
		"comm":      -16,
		"ip":        -3,
		"saddr":     -16,
		"daddr":     -16,
		"dport":     -7,
	}

	defaultColumns := []string{
		"node",
		"namespace",
		"pod",
		"container",
		"pid",
		"comm",
		"ip",
		"saddr",
		"daddr",
		"dport",
	}

	cmd := &cobra.Command{
		Use:   "tcpconnect",
		Short: "Trace connect system calls",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &utils.TraceConfig{
				GadgetName:       "tcpconnect",
				Operation:        "start",
				TraceOutputMode:  "Stream",
				TraceOutputState: "Started",
				CommonFlags:      &commonFlags,
			}

			// print header
			var requestedColumns []string
			switch commonFlags.OutputMode {
			case utils.OutputModeJSON:
				// Nothing to print
			case utils.OutputModeColumns:
				requestedColumns = defaultColumns
			case utils.OutputModeCustomColumns:
				requestedColumns = commonFlags.CustomColumns
			}
			printColumnsHeader(columnsWidth, requestedColumns)

			transformEvent := func(line string) string {
				var e types.Event

				if err := json.Unmarshal([]byte(line), &e); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %s", utils.WrapInErrUnmarshalOutput(err, line))
					return ""
				}

				if e.Type != eventtypes.NORMAL {
					utils.ManageSpecialEvent(e.Event, commonFlags.Verbose)
					return ""
				}

				return tcpconnectTransformEvent(e, columnsWidth, requestedColumns)
			}

			if err := utils.RunTraceAndPrintStream(config, transformEvent); err != nil {
				return utils.WrapInErrRunGadget(err)
			}

			return nil
		},
	}

	utils.AddCommonFlags(cmd, &commonFlags)

	return cmd
}

// tcpconnectTransformEvent is called to transform an event to columns format.
func tcpconnectTransformEvent(event types.Event, columnsWidth map[string]int, requestedColumns []string) string {
	var sb strings.Builder

	for _, col := range requestedColumns {
		switch col {
		case "node":
			sb.WriteString(fmt.Sprintf("%*s", columnsWidth[col], event.Node))
		case "namespace":
			sb.WriteString(fmt.Sprintf("%*s", columnsWidth[col], event.Namespace))
		case "pod":
			sb.WriteString(fmt.Sprintf("%*s", columnsWidth[col], event.Pod))
		case "container":
			sb.WriteString(fmt.Sprintf("%*s", columnsWidth[col], event.Container))
		case "pid":
			sb.WriteString(fmt.Sprintf("%*d", columnsWidth[col], event.Pid))
		case "comm":
			sb.WriteString(fmt.Sprintf("%*s", columnsWidth[col], event.Comm))
		case "ip":
			sb.WriteString(fmt.Sprintf("%*d", columnsWidth[col], event.IPVersion))
		case "saddr":
			sb.WriteString(fmt.Sprintf("%*s", columnsWidth[col], event.Saddr))
		case "daddr":
			sb.WriteString(fmt.Sprintf("%*s", columnsWidth[col], event.Daddr))
		case "dport":
			sb.WriteString(fmt.Sprintf("%*d", columnsWidth[col], event.Dport))
		}

		// Needed when field is larger than the predefined columnsWidth.
		sb.WriteRune(' ')
	}

	return sb.String()
}
