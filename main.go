package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/pkg/errors"

	"github.com/mwasilew2/openfga-rebac-example/services"
)

var (
	fgaApiUrl  = "http://:8080"
	fgaStoreID = "01JEQZ1TBAEXFDHKPZSZ3GNZG6"
	modelID    = "01JEQZ8CG0XYAN76DXZMM3RXTZ"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))
	slog.SetDefault(logger)

	var authzService services.AuthorizationService
	authzService, err := services.NewOpenFGAService(logger, fgaApiUrl, fgaStoreID, modelID)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to create OpenFGA service"))
	}

	// OpenFGA usage example
	func() {
		// remove Olive's access to secret
		err = authzService.DeleteUserAccess(context.Background(), "olive", "doc:secret1", "viewer")
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to delete Olive's relationship to secret"))
		}

		// expected to succeed, Jack should be able to read
		canRead, err := authzService.ReadUserAccess(context.Background(), "jack", "doc:secret1", "can_read")
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to check Jack's access to secret"))
		}
		fmt.Println("Can Jack read secret?", canRead)

		// expected to fail, Olive should not be able to read
		canRead, err = authzService.ReadUserAccess(context.Background(), "olive", "doc:secret1", "can_read")
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to check Olive's access to secret"))
		}
		fmt.Println("Can Olive read secret?", canRead)

		// update relationship to give Olive access to secret
		err = authzService.CreateUserAccess(context.Background(), "olive", "doc:secret1", "viewer")
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to create Olive's relationship to secret"))
		}

		// expected to succeed, Olive should be able to read
		canRead, err = authzService.ReadUserAccess(context.Background(), "olive", "doc:secret1", "can_read")
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to check Olive's access to secret"))
		}
		fmt.Println("Can Olive read secret?", canRead)
	}()
}
