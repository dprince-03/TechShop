"""Each service owns its database: the recommender may use ITS OWN Postgres only.

It must never be pointed at the Go backend's main database. Data from Go arrives through the
internal API; results go back through the recommender's API (a Go job pulls them).
"""
import importlib.util
import os

# Clients for other engines have no place in the recommender.
FORBIDDEN_DRIVERS = ["pymongo", "redis", "asyncpg", "sqlalchemy"]


def test_main_database_url_not_set() -> None:
    # The Go backend's connection string; the recommender uses RECS_DATABASE_URL instead.
    assert "DATABASE_URL" not in os.environ


def test_only_its_own_database_driver() -> None:
    found = [name for name in FORBIDDEN_DRIVERS if importlib.util.find_spec(name) is not None]
    assert not found, f"unexpected database drivers installed: {found}"


def test_recs_database_is_not_the_main_one() -> None:
    url = os.environ.get("RECS_DATABASE_URL", "")
    assert "@postgres:" not in url, "RECS_DATABASE_URL must point at recs-postgres, not the main database"
