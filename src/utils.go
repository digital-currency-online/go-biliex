package main

import (
	log "github.com/sirupsen/logrus"
)

func normalErr(err error, msg string) {
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error(msg)
	}
}

func fatalErr(err error, msg string) {
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal(msg)
	}
}
