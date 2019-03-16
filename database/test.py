from pymongo import MongoClient
import uuid


def sanity():
    client = MongoClient()
    db = client.test_database
    posts = db.posts
    entry = {"key1": uuid.uuid4(), "key2": uuid.uuid4()}
    posts.insert_one(entry)
    result = db.posts.find_one({"key1": entry["key1"]})
    assert result["key1"] == entry["key1"]
    assert result["key2"] == entry["key2"]


if __name__ == "__main__":
    sanity()
