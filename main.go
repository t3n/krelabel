package main

import (
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	zerolog.TimeFieldFormat = ""
	zerolog.LevelFieldName = "severity"
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal().Err(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err)
	}

	podName := os.Getenv("HOSTNAME")
	podNamespace := os.Getenv("NAMESPACE")
	podIP := os.Getenv("POD_IP")
	masterGroup := os.Getenv("MASTER_GROUP")
	patchSlave := []byte(`{"metadata":{"labels": {"role": "slave"}}}`)
	patchMaster := []byte(`{"metadata":{"labels": {"role": "master"}}}`)

	_, err = clientset.CoreV1().Pods(podNamespace).Patch(podName, types.MergePatchType, patchSlave)
	if err != nil {
		log.Fatal().Err(err)
	}
	log.Info().Str("role", "slave").Msg("set role")

	client := redis.NewSentinelClient(&redis.Options{
		Addr: "localhost:26379",
	})
	defer client.Close()

	go func() {
		for {
			masterIP := client.GetMasterAddrByName(masterGroup).Val()[0]

			if podIP == masterIP {
				_, err = clientset.CoreV1().Pods(podNamespace).Patch(podName, types.MergePatchType, patchMaster)
				if err != nil {
					log.Error().Err(err)
				}
				log.Info().Str("role", "master").Msg("set role")
			} else {
				_, err = clientset.CoreV1().Pods(podNamespace).Patch(podName, types.MergePatchType, patchSlave)
				if err != nil {
					log.Error().Err(err)
				}
				log.Info().Str("role", "slave").Msg("set role")
			}

			time.Sleep(30 * time.Second)
		}
	}()

	for {
		pubsub := client.PSubscribe("+switch-master")

		// Wait for confirmation that subscription is created before publishing anything.
		_, err := pubsub.Receive()
		if err != nil {
			log.Fatal().Err(err)
		}

		// Go channel which receives messages.
		ch := pubsub.Channel()

		time.AfterFunc(time.Second, func() {
			// When pubsub is closed channel is closed too.
			_ = pubsub.Close()
		})

		// Consume messages.
		for msg := range ch {
			log.Info().Str("channel", msg.Channel).Str("payload", msg.Payload).Msg("message")
			p := strings.Split(msg.Payload, " ")
			if p[0] == masterGroup {
				if p[3] == podIP {
					_, err = clientset.CoreV1().Pods(podNamespace).Patch(podName, types.MergePatchType, patchMaster)
					if err != nil {
						log.Error().Err(err)
					}
					log.Info().Str("role", "master").Msg("set role")
				} else {
					_, err = clientset.CoreV1().Pods(podNamespace).Patch(podName, types.MergePatchType, patchSlave)
					if err != nil {
						log.Error().Err(err)
					}
					log.Info().Str("role", "slave").Msg("set role")
				}
			}
		}
	}
}
