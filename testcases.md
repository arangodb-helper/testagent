Testcases for the test scripts
==============================

This document will grow over time. Each test case must describe

 1. what one-time initialisation is needed
 2. what operation to do
 3. which responses to treat as temporary errors
 4. what retry and backoff strategy to use
 5. which responses to treat as immediate failures

The idea is that tests run continuously and for the whole duration of the
test. They can be stopped temporarily and resumed later.

Read known documents
--------------------

 1. Initially, use

        POST /_api/collection

    with body

        { "name": "readtest", "numberOfShards": 3, "replicationFactor": 2 }

    to create a collection and write 100000 documents using

        POST /_api/document/readtest

    with body

        { "_key": "doc00000", "value": 0, "astring": "abc", "good": true }

    and similarly for the other 9999.

 2. Continuously read a document via

        GET /_api/document/readtest/doc00000

    varying the number between 00000 and 99999 in some way. Use any of the
    coordinators and switch from one to another continuously.

    Expect to get HTTP 200 OK.

 3. Connection failures and other HTTP error conditions should lead to
    retry (on a different coordinator). If an HTTP 200 OK can be achieved
    within 1 min the test script can ignore the issue (maybe with a log
    entry which counts as harmless).

 4. Simple retry different coordinators for a minute.

 5. A HTTP 404 "not found" or 307 "redirect" is an immediate failure.
    Failure to get an HTTP 200 within a minute is a test failure.

