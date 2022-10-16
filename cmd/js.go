/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"loader/pkg/delivery"
	"loader/pkg/domain"
	"loader/pkg/repository"
	"loader/pkg/service"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var natsServer, mongoServer, streams, creds string

// jsCmd represents the js command
var jsCmd = &cobra.Command{
	Use:   "js",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(log.Ldate | log.Lmicroseconds)
		streamList := strings.Split(streams, ",")

		wg := sync.WaitGroup{}
		wg.Add(len(streamList))

		router := gin.Default()

		ch := make(chan domain.Chan, len(streamList))

		auth := options.Credential{
			Username: os.Getenv("MGO_USER"),
			Password: os.Getenv("MGO_PASS"),
		}
		mgo, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", mongoServer)).SetConnectTimeout(2*time.Second).SetAuth(auth))
		if err != nil {
			log.Fatal("[ERR] connect database fail, ", err)
		}

		// nc, err := nats.Connect(natsServer, nats.UserCredentials(creds), nats.Name(fmt.Sprintf("nats-loader(%v)", streams)))
		nc, err := nats.Connect(natsServer, nats.Name(fmt.Sprintf("nats-loader(%v)", streams)))
		if err != nil {
			log.Fatal("[ERR] connect NATS server fail, ", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		for i := 0; i < len(streamList); i++ {
			go func(n int) {
				r := repository.NewLoaderRepo(mgo, nc, streamList[n])
				s := service.NewLoaderService(&r, streamList[n])

				c := domain.Chan{ChanService: s, Stream: streamList[n]}
				ch <- c

				s.Loader(streamList[n], ctx)
			}(i)
		}

		h := delivery.NewLoaderHandler(ch, len(streamList), cancel)

		h.Route(router)
		router.Run(":8890")

		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(jsCmd)

	jsCmd.Flags().StringVarP(&natsServer, "source", "s", "localhost:4222", "")
	jsCmd.Flags().StringVarP(&mongoServer, "destination", "d", "localhost:27017", "")
	jsCmd.Flags().StringVarP(&streams, "streams", "t", "", "")
	jsCmd.Flags().StringVarP(&creds, "creds", "c", "", "")
}
