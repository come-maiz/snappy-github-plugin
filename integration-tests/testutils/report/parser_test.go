// -*- Mode: Go; indent-tabs-mode: t -*-
// +build !excludeintegration

/*
 * Copyright (C) 2015 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package report

import (
	"bytes"
	"fmt"

	"github.com/testing-cabal/subunit-go"
	"gopkg.in/check.v1"
)

var _ = check.Suite(&ParserReportSuite{})

type StatuserSpy struct {
	calls []subunit.Event
}

func (s *StatuserSpy) Status(event subunit.Event) error {
	s.calls = append(s.calls, event)
	return nil
}

type ParserReportSuite struct {
	subject *SubunitV2ParserReporter
	spy     *StatuserSpy
	output  bytes.Buffer
}

func (s *ParserReportSuite) SetUpTest(c *check.C) {
	s.spy = &StatuserSpy{}
	s.subject = &SubunitV2ParserReporter{statuser: s.spy}
}

func (s *ParserReportSuite) TestParserSendsNothingWitNotParseableInput(c *check.C) {
	s.subject.Write([]byte("Not parseable"))

	c.Assert(len(s.spy.calls), check.Equals, 0,
		check.Commentf("Unexpected event sent to subunit: %v", s.spy.calls))
}

var eventTests = []struct {
	gocheckOutput  string
	expectedTestID string
	expectedStatus string
}{{
	"****** Running testSuite.TestExists\n",
	"testSuite.TestExists",
	"exists",
}, {
	"PASS: /tmp/snappy-tests-job/18811/src/github.com/ubuntu-core/snappy/integration-tests/tests/" +
		"apt_test.go:34: testSuite.TestSuccess      0.005s\n",
	"testSuite.TestSuccess",
	"success",
}, {
	"FAIL: /tmp/snappy-tests-job/710/src/github.com/ubuntu-core/snappy/integration-tests/tests/" +
		"installFramework_test.go:85: testSuite.TestFail\n",
	"testSuite.TestFail",
	"fail",
}}

func (s *ParserReportSuite) TestParserReporterSendsEvents(c *check.C) {
	for _, t := range eventTests {
		s.spy.calls = []subunit.Event{}
		s.subject.Write([]byte(t.gocheckOutput))

		c.Check(s.spy.calls, check.HasLen, 1)
		event := s.spy.calls[0]
		c.Check(event.TestID, check.Equals, t.expectedTestID)
		c.Check(event.Status, check.Equals, t.expectedStatus)
	}
}

func (s *ParserReportSuite) TestParserReporterSendsSkipEvent(c *check.C) {
	testID := "testSuite.TestSkip"
	skipReason := "skip reason"
	s.subject.Write([]byte(
		fmt.Sprintf("SKIP: /tmp/snappy-tests-job/21647/src/github.com/ubuntu-core/snappy/"+
			"integration-tests/tests/info_test.go:36: %s (%s)\n", testID, skipReason)))

	c.Check(s.spy.calls, check.HasLen, 1)
	event := s.spy.calls[0]
	c.Check(event.TestID, check.Equals, testID)
	c.Check(event.Status, check.Equals, "skip")
	c.Check(event.MIME, check.Equals, "text/plain;charset=utf8")
	c.Check(event.FileName, check.Equals, "reason")
	c.Check(string(event.FileBytes), check.Equals, skipReason)
}
