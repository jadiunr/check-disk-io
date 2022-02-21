package main

import (
	"fmt"
	"strings"

	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-plugin-sdk/sensu"
	"github.com/shirou/gopsutil/v3/disk"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
}

type MetricGroup struct {
	Comment string
	Type 	string
	Name 	string
	Metrics []Metric
}

func (g *MetricGroup) AddMetric(tags map[string]string, value float64) {
	g.Metrics = append(g.Metrics, Metric{
		Tags:	tags,
		Value: 	value,
	})
}

func (g *MetricGroup) Output() {
	var output string
	fmt.Printf("# HELP %s [%s] %s\n", g.Name, g.Type, g.Comment)
	fmt.Printf("# TYPE %s %s\n", g.Name, g.Type)
	for _, m := range g.Metrics {
		tagStr := ""
		for tag, tvalue := range m.Tags {
			if len(tagStr) > 0 {
				tagStr = tagStr + ","
			}
			tagStr = tagStr + tag + "=\"" + tvalue + "\""
		}
		if len(tagStr) > 0 {
			tagStr = "{" + tagStr + "}"
		}
		output = strings.Join([]string{g.Name + tagStr, fmt.Sprintf("%v", m.Value)}, " ")
		fmt.Println(output)
	}
	fmt.Println("")
}

type Metric struct {
	Tags 	map[string]string
	Value 	float64
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "check-disk-io",
			Short:    "Check disk IO and provide metrics",
			Keyspace: "sensu.io/plugins/check-disk-io/config",
		},
	}
)

func main() {
	check := sensu.NewGoCheck(&plugin.PluginConfig, nil, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *types.Event) (int, error) {
	return sensu.CheckStateOK, nil
}

func executeCheck(event *types.Event) (int, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		fmt.Printf("Failed to get partitions, error: %v", err)
	}

	metricGroups := map[string]*MetricGroup{
		"disk_read_bytes": {
			Name: "disk_read_bytes",
			Type: "COUNTER",
			Comment: "These values count the number of bytes read from or written to this block device.",
		},
		"disk_write_bytes": {
			Name: "disk_write_bytes",
			Type: "COUNTER",
			Comment: "These values count the number of bytes read from or written to this block device.",
		},
		"disk_read_count": {
			Name: "disk_read_count",
			Type: "COUNTER",
			Comment: "These values increment when an I/O request completes.",
		},
		"disk_write_count": {
			Name: "disk_write_count",
			Type: "COUNTER",
			Comment: "These values increment when an I/O request completes.",
		},
		"disk_read_time": {
			Name: "disk_read_time",
			Type: "COUNTER",
			Comment: "These values count the number of milliseconds that I/O requests have waited on this block device. If there are multiple I/O requests waiting, these values will increase at a rate greater than 1000/second; for example, if 60 read requests wait for an average of 30 ms, the read_time field will increase by 60*30 = 1800.",
		},
		"disk_write_time": {
			Name: "disk_write_time",
			Type: "COUNTER",
			Comment: "These values count the number of milliseconds that I/O requests have waited on this block device. If there are multiple I/O requests waiting, these values will increase at a rate greater than 1000/second; for example, if 60 read requests wait for an average of 30 ms, the read_time field will increase by 60*30 = 1800.",
		},
		"disk_io_time": {
			Name: "disk_io_time",
			Type: "COUNTER",
			Comment: "This value counts the number of milliseconds during which the device has had I/O requests queued.",
		},
		"disk_weighted_io": {
			Name: "disk_weighted_io",
			Type: "COUNTER",
			Comment: "This value counts the number of milliseconds that I/O requests have waited on this block device. If there are multiple I/O requests waiting, this value will increase as the product of the number of milliseconds times the number of requests waiting (see disk_read_time for an example).",
		},
		"disk_iops_in_progress": {
			Name: "disk_iops_in_progress",
			Type: "GAUGE",
			Comment: "This value counts the number of I/O requests that have been issued to the device driver but have not yet completed. It does not include I/O requests that are in the queue but not yet issued to the device driver.",
		},
		"disk_merged_read_count": {
			Name: "disk_merged_read_count",
			Type: "COUNTER",
			Comment: "Reads and writes which are adjacent to each other may be merged for efficiency. Thus, two 4K reads may become one 8K read before it is ultimately handed to the disk, and so it will be counted (and queued) as only one I/O. These fields lets you know how often this was done.",
		},
		"disk_merged_write_count": {
			Name: "disk_merged_write_count",
			Type: "COUNTER",
			Comment: "Reads and writes which are adjacent to each other may be merged for efficiency. Thus, two 4K reads may become one 8K read before it is ultimately handed to the disk, and so it will be counted (and queued) as only one I/O. These fields lets you know how often this was done.",
		},
	}

	for _, p := range parts {
		diskio, err := disk.IOCounters(p.Device)
		if err != nil {
			fmt.Printf("Failed to get IO counters, error: %v", err)
		}
		for _, v := range diskio {
			tags := map[string]string{"device": v.Name, "mountpoint": p.Mountpoint}
			metricGroups["disk_read_bytes"].AddMetric(tags, float64(v.ReadBytes))
			metricGroups["disk_write_bytes"].AddMetric(tags, float64(v.WriteBytes))
			metricGroups["disk_read_count"].AddMetric(tags, float64(v.ReadCount))
			metricGroups["disk_write_count"].AddMetric(tags, float64(v.WriteCount))
			metricGroups["disk_read_time"].AddMetric(tags, float64(v.ReadTime))
			metricGroups["disk_write_time"].AddMetric(tags, float64(v.WriteTime))
			metricGroups["disk_io_time"].AddMetric(tags, float64(v.IoTime))
			metricGroups["disk_weighted_io"].AddMetric(tags, float64(v.WeightedIO))
			metricGroups["disk_iops_in_progress"].AddMetric(tags, float64(v.IopsInProgress))
			metricGroups["disk_merged_read_count"].AddMetric(tags, float64(v.MergedReadCount))
			metricGroups["disk_merged_write_count"].AddMetric(tags, float64(v.MergedWriteCount))
		}
	}

	for _, v := range metricGroups {
		v.Output()
	}

	return sensu.CheckStateOK, nil
}
