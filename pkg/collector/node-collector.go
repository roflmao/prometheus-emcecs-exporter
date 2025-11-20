package collector

import (
	"github.com/paychex/prometheus-emcecs-exporter/pkg/ecsclient"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// A EcsNodeDTCollector implements the prometheus.Collector.
type EcsNodeDTCollector struct {
	ecsClient *ecsclient.EcsClient
	namespace string
}

var (
	// Disk metrics
	numDisks = prometheus.NewDesc(
		prometheus.BuildFQName("emcecs", "node", "disks_total"),
		"Total number of disks on node",
		[]string{"node"}, nil,
	)
	numGoodDisks = prometheus.NewDesc(
		prometheus.BuildFQName("emcecs", "node", "disks_good"),
		"Number of good disks on node",
		[]string{"node"}, nil,
	)
	numBadDisks = prometheus.NewDesc(
		prometheus.BuildFQName("emcecs", "node", "disks_bad"),
		"Number of bad disks on node",
		[]string{"node"}, nil,
	)
	// Storage metrics
	diskSpaceTotal = prometheus.NewDesc(
		prometheus.BuildFQName("emcecs", "node", "disk_space_total_bytes"),
		"Total disk space on node in bytes",
		[]string{"node"}, nil,
	)
	diskSpaceFree = prometheus.NewDesc(
		prometheus.BuildFQName("emcecs", "node", "disk_space_free_bytes"),
		"Free disk space on node in bytes",
		[]string{"node"}, nil,
	)
	diskSpaceAllocated = prometheus.NewDesc(
		prometheus.BuildFQName("emcecs", "node", "disk_space_allocated_bytes"),
		"Allocated disk space on node in bytes",
		[]string{"node"}, nil,
	)
	// Active connections
	activeConnections = prometheus.NewDesc(
		prometheus.BuildFQName("emcecs", "node", "active_connections"),
		"Number of current active connections on node",
		[]string{"node"}, nil,
	)
	// NOTE: ObjectScale 4.1 Dashboard API does not provide CPU, memory, network, or transaction metrics
)

// NewEcsNodeDTCollector returns an initialized Node DT Collector.
func NewEcsNodeDTCollector(emcecs *ecsclient.EcsClient, namespace string) (*EcsNodeDTCollector, error) {

	log.WithFields(log.Fields{"package": "node-collector"}).Debug("Init Node exporter")
	return &EcsNodeDTCollector{
		ecsClient: emcecs,
		namespace: namespace,
	}, nil
}

// Collect fetches the stats from configured nodes as Prometheus metrics.
// It implements prometheus.Collector.
func (e *EcsNodeDTCollector) Collect(ch chan<- prometheus.Metric) {
	log.WithFields(log.Fields{"package": "node-collector"}).Debug("ECS Node collect starting")
	if e.ecsClient == nil {
		log.WithFields(log.Fields{"package": "node-collector"}).Error("ECS client not configured.")
		return
	}

	nodeState := e.ecsClient.RetrieveNodeStateParallel()
	for _, node := range nodeState {
		// Disk metrics
		ch <- prometheus.MustNewConstMetric(numDisks, prometheus.GaugeValue, node.NumDisks, node.NodeIP)
		ch <- prometheus.MustNewConstMetric(numGoodDisks, prometheus.GaugeValue, node.NumGoodDisks, node.NodeIP)
		ch <- prometheus.MustNewConstMetric(numBadDisks, prometheus.GaugeValue, node.NumBadDisks, node.NodeIP)

		// Storage metrics
		ch <- prometheus.MustNewConstMetric(diskSpaceTotal, prometheus.GaugeValue, node.DiskSpaceTotal, node.NodeIP)
		ch <- prometheus.MustNewConstMetric(diskSpaceFree, prometheus.GaugeValue, node.DiskSpaceFree, node.NodeIP)
		ch <- prometheus.MustNewConstMetric(diskSpaceAllocated, prometheus.GaugeValue, node.DiskSpaceAllocated, node.NodeIP)

		// Active connections
		ch <- prometheus.MustNewConstMetric(activeConnections, prometheus.GaugeValue, node.ActiveConnections, node.NodeIP)
	}

	log.WithFields(log.Fields{"package": "node-collector"}).Debug("Node exporter finished")
	log.WithFields(log.Fields{"package": "node-collector"}).Debug(nodeState)
}

// Describe describes the metrics exported from this collector.
func (e *EcsNodeDTCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- numDisks
	ch <- numGoodDisks
	ch <- numBadDisks
	ch <- diskSpaceTotal
	ch <- diskSpaceFree
	ch <- diskSpaceAllocated
	ch <- activeConnections
}
