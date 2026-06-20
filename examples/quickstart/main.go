// Command quickstart demonstrates the CrawlSnap Go SDK.
//
//	export CRAWLSNAP_API_KEY=sk-cs-...
//	go run ./examples/quickstart
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	crawlsnap "github.com/crawlsnap/crawlsnap-go"
)

func main() {
	// API key falls back to $CRAWLSNAP_API_KEY when the first argument is empty.
	client, err := crawlsnap.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// VectorSnap — IoC reputation enrichment.
	ip, err := client.VectorSnap.IP(ctx, "8.8.8.8")
	if err != nil {
		var nf *crawlsnap.NotFoundError
		if errors.As(err, &nf) {
			fmt.Println("no data for that indicator")
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("reputation=%d as_owner=%s\n", ip.GetReputation(), ip.GetAsOwner())
	}

	// Pin a single product to a specific API version (others unaffected).
	if _, err := client.PulseSnap.V1().Domain(ctx, "example.com"); err != nil {
		log.Println("pulse:", err)
	}

	// SubdoSnap — stream every subdomain across all pages.
	for sub, err := range client.SubdoSnap.ScanIter(ctx, "example.com") {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(sub["subdomain"])
	}
}
