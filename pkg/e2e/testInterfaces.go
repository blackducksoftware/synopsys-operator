package e2e

import (
	log "github.com/sirupsen/logrus"
)

type TestSuiteInterface interface {
	GetTests() []TestInterface
	Setup()
	Cleanup()
}

type TestInterface interface {
	GetName() string
	TestToRun() error
	Cleanup() []error
}

// RunTestSuite TODO
func RunTestSuite(testSuite TestSuiteInterface) {
	for _, test := range testSuite.GetTests() {
		testErr := test.TestToRun()
		cleanupErr := test.Cleanup()
		if testErr != nil || len(cleanupErr) != 0 {
			log.Errorf("test %v failed with testErr: %v and cleanupErr: %v", test.GetName(), testErr, cleanupErr)
		}
	}
	testSuite.Cleanup()
}

// RunTestSuites TODO
func RunTestSuites(testSuites []TestSuiteInterface) {
	for _, testSuite := range testSuites {
		RunTestSuite(testSuite)
	}
}
