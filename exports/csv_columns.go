package exports

import (
	"encoding/csv"
	"fmt"
	"github.com/boredomist/mixport/mixpanel"
	"io"
)

// EventDef represents the definition of an event's CSV columns to be passed on
// to the `CSVColumnStreamer` function.
//
// XXX: This is a pretty non-descript name, it should be changed.
type EventDef struct {
	writer  *csv.Writer
	columns []string
}

// NewEventDef oddly enough creates an instance of the EventDef struct from the
// given io.Writer and list of column names.
//
// Columns in the output will be in the same order as they passed in here.
func NewEventDef(w io.Writer, columns []string) EventDef {
	return EventDef{
		writer:  csv.NewWriter(w),
		columns: columns,
	}
}

// CSVColumnStreamer writes CSVs with explicitly defined events and
// properties. This is useful if only a subset of the properties attached to an
// event type are useful or the data needs to be stored in a traditional SQL
// table with columns known ahead of time.
//
// This will write to a unique io.Writer for each specified event.
//
// The `defs` map contains a mapping of the event names to capture to their
// EventDefs. Any event received that is not in this map will simply be dropped.
func CSVColumnStreamer(defs map[string]EventDef, records <-chan mixpanel.EventData) {

	for _, def := range defs {
		// Write the column names as CSV header
		def.writer.Write(def.columns)
	}

	for record := range records {
		event := record["event"].(string)

		// We simply ignore events we don't have column definitions
		// for.
		if def, ok := defs[event]; ok {
			cols := make([]string, len(def.columns))

			// If the property is nil or doesn't exist in the event
			// data, assign it an empty string value.
			for i, col := range def.columns {
				switch value := record[col]; value.(type) {
				case nil:
					cols[i] = ""
				default:
					cols[i] = fmt.Sprintf("%v", value)
				}
			}

			def.writer.Write(cols)
		}
	}

	// Flush any remaining buffered data to the underlying io.Writer
	for _, def := range defs {
		def.writer.Flush()
	}
}