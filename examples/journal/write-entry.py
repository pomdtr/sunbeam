#!/usr/bin/env python3

import argparse
from journal import JournalEntry, load_journal, save_journal

parser = argparse.ArgumentParser()
parser.add_argument('--title')
parser.add_argument('--message')

args = parser.parse_args()

entry: JournalEntry = {
    'title': args.title,
    'message': args.message,
    'date': '2021-01-01',
}

journal = load_journal()
journal['entries'].append(entry)
save_journal(journal)
