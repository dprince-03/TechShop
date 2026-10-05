"""Typed settings from the environment. recs_database_url is the recommender's OWN database only."""
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    techshop_internal_url: str
    recs_client_id: str
    recs_client_secret: str
    recs_database_url: str
    artifact_dir: str = "./artifacts"
