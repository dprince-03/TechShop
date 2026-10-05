"""Command line: techshop-recs run-nightly | train <model> | evaluate | sample."""
import typer

app = typer.Typer(help="TechShop recommendations")


@app.command("run-nightly")
def run_nightly() -> None:
    """Export → train gated models → evaluate → import (pipelines/nightly.py)."""
    raise NotImplementedError("scaffold")


@app.command()
def train(model: str) -> None:
    """Train a single generator by name, e.g. `popularity`."""
    raise NotImplementedError("scaffold")


@app.command()
def evaluate() -> None:
    """Score the latest run against the baselines and write a report."""
    raise NotImplementedError("scaffold")
