package main

import (
	"flag"
	"fmt"
	"os"

	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	protorekor "github.com/sigstore/protobuf-specs/gen/pb-go/rekor/v1"
	"github.com/sigstore/rekor/pkg/client"
	"github.com/sigstore/rekor/pkg/generated/client/entries"
	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/sigstore/rekor/pkg/tle"
	bundle "github.com/sigstore/sigstore-go/pkg/bundle"
	"golang.org/x/mod/semver"
	"google.golang.org/protobuf/encoding/protojson"
)

var version string
var bundlePath string
var pretty bool

func init() {
	flag.Usage = func() {
		println("Usage: sigstore-bundle-upgrade <path/to/sigstore/bundle>")
		flag.PrintDefaults()
	}
	flag.BoolVar(&pretty, "pretty", false, "Pretty print the output")
	flag.StringVar(&version, "version", "0.3", "Bundle version to upgrade to")
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	bundlePath = flag.Arg(0)
}

func main() {
	if err := runConvert(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runConvert() error {
	b, err := bundle.LoadJSONFromPath(bundlePath)
	if err != nil {
		return err
	}
	mediaType, err := bundle.MediaTypeString(version)
	if err != nil {
		return err
	}
	if semver.Compare("v"+version, "v0.2") >= 0 {
		if b.Bundle.VerificationMaterial.TlogEntries != nil {
			for i := 0; i < len(b.Bundle.VerificationMaterial.TlogEntries); i++ {
				if b.Bundle.VerificationMaterial.TlogEntries[i].InclusionProof == nil {
					b.Bundle.VerificationMaterial.TlogEntries[i], err = convertTLogEntry(b.Bundle.VerificationMaterial.TlogEntries[i])
					if err != nil {
						return err
					}
				}
			}
		}
	}
	if semver.Compare("v"+version, "v0.3") >= 0 {
		if certChain, ok := b.Bundle.VerificationMaterial.GetContent().(*protobundle.VerificationMaterial_X509CertificateChain); ok {
			if len(certChain.X509CertificateChain.Certificates) == 0 {
				return fmt.Errorf("certificate chain empty")
			}
			b.Bundle.VerificationMaterial.Content = &protobundle.VerificationMaterial_Certificate{
				Certificate: certChain.X509CertificateChain.Certificates[0],
			}
		}
	}
	b.Bundle.MediaType = mediaType
	marshalOptions := protojson.MarshalOptions{}
	if pretty {
		marshalOptions.Multiline = true
		marshalOptions.Indent = "\t"
	}
	outBytes, err := marshalOptions.Marshal(b.Bundle)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", outBytes)
	return nil
}

func convertTLogEntry(entry *protorekor.TransparencyLogEntry) (*protorekor.TransparencyLogEntry, error) {
	// TODO: support alternative transparency log servers
	rekor, err := client.GetRekorClient("https://rekor.sigstore.dev")
	if err != nil {
		return nil, err
	}
	logEntries, err := rekor.Entries.GetLogEntryByIndex(&entries.GetLogEntryByIndexParams{
		LogIndex: entry.LogIndex,
	})
	if err != nil {
		return nil, err
	}
	if len(logEntries.Payload) != 1 {
		return nil, fmt.Errorf("failed to retrieve inclusion proof")
	}
	var logEntry models.LogEntryAnon
	for _, entry := range logEntries.Payload {
		logEntry = entry
	}
	return tle.GenerateTransparencyLogEntry(logEntry)
}
