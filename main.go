// slurp s3 bucket enumerator
// Copyright (C) 2019 hehnope
//
// slurp is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// slurp is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Foobar. If not, see <http://www.gnu.org/licenses/>.
//

package main

import (
	"bufio"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	log "github.com/sirupsen/logrus"

	"slurp/scanner/cmd"
	"slurp/scanner/external"
	"slurp/scanner/intern"
	"slurp/scanner/stats"
)

// Global config
var cfg cmd.Config

func main() {
	cfg = cmd.Init("slurp", "Public buckets finder", "Public buckets finder")
	cfg.Stats = stats.NewStats()

	switch cfg.State {
	case "DOMAIN":
		for _, domain := range cfg.Domains {
			if !cfg.NoStats {
				cfg.Stats = stats.NewStats() // This will create a new stats instance, clearing the old one
			}
			cfg.Domains = []string{domain}
			external.Init(&cfg)

			log.Info("Building permutations....")
			go external.PermutateDomainRunner(&cfg)

			log.Info("Processing permutations....")
			external.CheckDomainPermutations(&cfg)

			// Print stats info
			log.Printf("%+v", cfg.Stats)
		}
	case "KEYWORD":
		external.Init(&cfg)

		log.Info("Building permutations....")
		go external.PermutateKeywordRunner(&cfg)

		log.Info("Processing permutations....")
		external.CheckKeywordPermutations(&cfg)

		// Print stats info
		log.Printf("%+v", cfg.Stats)
	case "INTERNAL":
		var config aws.Config
		config.Region = &cfg.Region

		log.Info("Determining public buckets....")
		buckets, err3 := intern.GetPublicBuckets(config)
		if err3 != nil {
			log.Error(err3)
		}

		for bucket := range buckets.ACL {
			log.Infof("S3 public bucket (ACL): %s", buckets.ACL[bucket])
		}

		for bucket := range buckets.Policy {
			log.Infof("S3 public bucket (Policy): %s", buckets.Policy[bucket])
		}
	case "DOMAINLIST":
		file, err := os.Open("domainlist.txt")
		if err != nil {
			log.Error("Error opening domainlist.txt:", err)
			os.Exit(1)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			domain := scanner.Text()

			if !cfg.NoStats {
				cfg.Stats = stats.NewStats() // This will create a new stats instance, clearing the old one
			}

			cmd := exec.Command("slurp", "domain", "-t", domain)
			output, err := cmd.Output()

			if err != nil {
				log.Error("Error running slurp command for domain:", domain, err)
				continue
			}

			log.Infof("Output for %s: %s", domain, string(output))
		}

		if err := scanner.Err(); err != nil {
			log.Error("Error reading from domainlist.txt:", err)
			os.Exit(1)
		}

	default:
		log.Fatal("Check help")
	}
}
