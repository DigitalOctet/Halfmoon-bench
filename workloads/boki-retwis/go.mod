module cs.utexas.edu/zjia/faas-retwis

go 1.14

require (
	cs.utexas.edu/zjia/faas v0.0.0
	cs.utexas.edu/zjia/faas/slib v0.0.0
	github.com/golang/snappy v0.0.2 // indirect
	github.com/montanaflynn/stats v0.6.3
	github.com/openacid/low v0.1.21 // indirect
	go.mongodb.org/mongo-driver v1.4.6
)

replace cs.utexas.edu/zjia/faas => ../../boki/worker/golang

replace cs.utexas.edu/zjia/faas/slib => ../../boki/slib
