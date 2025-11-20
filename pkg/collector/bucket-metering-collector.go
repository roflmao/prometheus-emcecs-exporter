package collector

import (
	"math"
	"time"

	"github.com/paychex/prometheus-emcecs-exporter/pkg/ecsclient"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// A EcsBucketMeteringCollector implements the prometheus.Collector.
type EcsBucketMeteringCollector struct {
	ecsClient *ecsclient.EcsClient
	namespace string
}

var (
	bucketObjectCount = prometheus.NewDesc(
		prometheus.BuildFQName("emcecs", "metering_bucket", "object_count"),
		"total count of objects in bucket",
		[]string{"namespace", "bucket"}, nil,
	)
	bucketSize = prometheus.NewDesc(
		prometheus.BuildFQName("emcecs", "metering_bucket", "size_bytes"),
		"total size of bucket in bytes",
		[]string{"namespace", "bucket"}, nil,
	)
	bucketIngressBytes = prometheus.NewDesc(
		prometheus.BuildFQName("emcecs", "metering_bucket", "ingress_bytes_total"),
		"total ingress bytes for bucket",
		[]string{"namespace", "bucket"}, nil,
	)
	bucketEgressBytes = prometheus.NewDesc(
		prometheus.BuildFQName("emcecs", "metering_bucket", "egress_bytes_total"),
		"total egress bytes for bucket",
		[]string{"namespace", "bucket"}, nil,
	)
)

// NewEcsBucketMeteringCollector returns an initialized Bucket Metering Collector.
func NewEcsBucketMeteringCollector(emcecs *ecsclient.EcsClient, namespace string) (*EcsBucketMeteringCollector, error) {

	log.WithFields(log.Fields{"package": "bucket-metering-collector"}).Debug("Init Bucket Metering exporter")
	return &EcsBucketMeteringCollector{
		ecsClient: emcecs,
		namespace: namespace,
	}, nil
}

// Collect fetches the bucket metering information from the cluster
// It implements prometheus.Collector.
func (e *EcsBucketMeteringCollector) Collect(ch chan<- prometheus.Metric) {
	log.WithFields(log.Fields{"package": "bucket-metering-collector"}).Debug("ECS Bucket Metering collect starting")
	if e.ecsClient == nil {
		log.WithFields(log.Fields{"package": "bucket-metering-collector"}).Error("ECS client not configured.")
		return
	}
	start := time.Now()

	// Get list of all namespaces
	nameSpaceReq := "https://" + e.ecsClient.ClusterAddress + ":4443/object/namespaces"
	n, err := e.ecsClient.CallECSAPI(nameSpaceReq)
	if err != nil {
		log.WithFields(log.Fields{"package": "bucket-metering-collector"}).Errorf("Error getting namespaces: %s", err)
		return
	}

	result := gjson.Get(n, "namespace.#.name")
	// We need to limit the number of requests going to the API at once
	// setting this to a max of 4 connections after testing
	concurrency := 4
	sem := make(chan bool, concurrency)
	gb2bytes := math.Pow10(9)

	for _, name := range result.Array() {
		// since we need this a few times, lets get the name once
		ns := name.String()

		// ensuring we don't overload the ECS with multiple calls, so we are limiting concurrency to 4
		sem <- true
		go func() {
			defer func() { <-sem }()

			// Retrieve bucket billing info for this namespace
			// POST request with JSON body to get all buckets
			bucketBillingReq := "https://" + e.ecsClient.ClusterAddress + ":4443/object/billing/buckets/" + ns + "/info"
			jsonBody := `{"bucketName":"*","sizeunit":"GB"}`

			n, err := e.ecsClient.CallECSAPIPost(bucketBillingReq, jsonBody)
			if err != nil {
				log.WithFields(log.Fields{"package": "bucket-metering-collector"}).Errorf("Error getting bucket billing info for namespace %s: %s", ns, err)
				return
			}

			// Parse bucket billing info
			// Response format: {"bucket_billing_info": [{"name": "bucket1", "total_size": 123, ...}, ...]}
			buckets := gjson.Get(n, "bucket_billing_info.#.name")

			for i, bucket := range buckets.Array() {
				bucketName := bucket.String()

				// Get metrics for this bucket from the array
				bucketInfo := gjson.Get(n, "bucket_billing_info."+string(rune(i)))

				totalObjects := gjson.Get(bucketInfo.String(), "total_objects").Float()
				totalSizeGB := gjson.Get(bucketInfo.String(), "total_size").Float()
				totalSizeBytes := totalSizeGB * gb2bytes
				ingressBytes := gjson.Get(bucketInfo.String(), "ingress_bytes").Float()
				egressBytes := gjson.Get(bucketInfo.String(), "egress_bytes").Float()

				ch <- prometheus.MustNewConstMetric(bucketObjectCount, prometheus.GaugeValue, totalObjects, ns, bucketName)
				ch <- prometheus.MustNewConstMetric(bucketSize, prometheus.GaugeValue, totalSizeBytes, ns, bucketName)
				ch <- prometheus.MustNewConstMetric(bucketIngressBytes, prometheus.CounterValue, ingressBytes, ns, bucketName)
				ch <- prometheus.MustNewConstMetric(bucketEgressBytes, prometheus.CounterValue, egressBytes, ns, bucketName)
			}
		}()
	}
	// This ensures that all our go routines completed
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	duration := float64(time.Since(start).Seconds())
	log.WithFields(log.Fields{"package": "bucket-metering-collector"}).Debugf("Scrape of bucket metering took %f seconds for cluster %s\n", duration, e.ecsClient.ClusterAddress)
	log.WithFields(log.Fields{"package": "bucket-metering-collector"}).Debug("Bucket Metering exporter finished")
}

// Describe describes the metrics exported from this collector.
func (e *EcsBucketMeteringCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- bucketObjectCount
	ch <- bucketSize
	ch <- bucketIngressBytes
	ch <- bucketEgressBytes
}
