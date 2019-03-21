import uuid

import fuzzer.database.results_db as results_db
import fuzzer.lib.exceptions as exceptions

RESULTS_DB = "athena"


def test_results_db():
    my_results_db = results_db.ResultsDb(RESULTS_DB)

    # Test insertion
    my_exception = exceptions.TargetException(
        "put",
        "/this/is/a/test/route",
        "NoMethodError",
        "This is a message from an exception",
    )
    my_exception_dict = my_exception.to_dict()

    my_results_db.write_exceptions([my_exception])
    result = my_results_db.find_exception_by_key({"verb": my_exception_dict["verb"]})
    assert result["message"] == my_exception_dict["message"]
    assert result["class"] == my_exception_dict["class"]
    assert result["path"] == my_exception_dict["path"]
    my_results_db.print_all_exceptions()


if __name__ == "__main__":
    test_results_db()
