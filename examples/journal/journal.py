from typing import TypedDict
import pathlib
import json

class JournalEntry(TypedDict):
    title: str
    content: str
    timestamp: int

class Journal(TypedDict):
    entries: dict[str, JournalEntry]

dirname = pathlib.Path(__file__).parent.absolute()

def load_journal() -> Journal:
    with open(dirname / 'journal.json') as f:
        journal: Journal = json.load(f)

    return journal

def save_journal(journal: Journal) -> None:
    with open(dirname / 'journal.json', 'w') as f:
        json.dump(journal, f, indent=4)
