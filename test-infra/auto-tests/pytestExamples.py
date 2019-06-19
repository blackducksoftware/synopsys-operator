import pytest
import logging

# Start Here...
'''
Running All Tests in File:
> pytest pytestExamples.py

Running a Specific Test:
> pytest pytestExamples.py::test_examplePass

Running a Test by Decorators:
> pytest -m <decorator>
> pytest -m exampleDecorator

Passing Arguments to Tests:
> 

'''


def test_examplePass():
    logging.info("TEST: test_examplePass")
    pass


@pytest.mark.exampleDecorator
def test_examplePassWithDecorator():
    logging.info("TEST: test_examplePassWithDecorator")
    pass


def test_examplePassWithArgs(args):
    logging.info("TEST: test_examplePassWithArgs")
    pass


def test_exampleAssertFail():
    logging.info("TEST: test_exampleAssertFail")
    assert 0


def test_exampleRaiseException():
    logging.info("TEST: test_exampleRaiseException")
    raise Exception("Example Raise")


def test_examplePytestFail():
    logging.info("TEST: test_examplePytestFail")
    pytest.fail("test_examplePytestFail")
