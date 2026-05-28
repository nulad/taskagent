import json
from typing import Any


def print_json(data: Any) -> None:
    """Print data as JSON to stdout."""
    json.dump(data, fp=__import__("sys").stdout)
    __import__("sys").stdout.write("\n")
