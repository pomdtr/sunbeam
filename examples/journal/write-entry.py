#!/usr/bin/env python3

import argparse
from journal import JournalEntry, load_journal, save_journal
from datetime import datetime
from uuid import uuid4

parser = argparse.ArgumentParser()
parser.add_argument('--title')
parser.add_argument('--content')

args = parser.parse_args()

entry: JournalEntry = {
    'title': args.title,
    'content': args.content,
    'timestamp': int(datetime.now().timestamp()),
}

journal = load_journal()
uuid = str(uuid4())
journal['entries'][uuid] = entry
save_journal(journal)
