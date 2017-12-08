package influxdb

import (
	"github.com/aptible/mini-collector/batch"
	client "github.com/influxdata/influxdb/client/v2"
	log "github.com/sirupsen/logrus"
)

func entryToFields(entry *batch.Entry) map[string]interface{} {
	return map[string]interface{}{
		// NOTE: Older versions of InfluxDB do not support uint64 here.
		"milli_cpu_usage": int64((*entry).MilliCpuUsage),
		"memory_total_mb": int64((*entry).MemoryTotalMb),
		"memory_rss_mb":   int64((*entry).MemoryRssMb),
		"memory_limit_mb": int64((*entry).MemoryLimitMb),
		"running":         (*entry).Running,
	}
}

func buildBatchPoints(database string, batch []batch.Entry) client.BatchPoints {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})

	if err != nil {
		// These are just fatal errors because they're caused
		// by coding errors, not transient problems.
		log.Fatalf("could not build batch points: %+v", err)
	}

	for _, entry := range batch {
		fields := entryToFields(&entry)

		pt, err := client.NewPoint("metrics", entry.Tags, fields, entry.Time)
		if err != nil {
			// Same as above
			log.Fatalf("could not build point: %+v", err)
		}

		bp.AddPoint(pt)
	}

	return bp
}
