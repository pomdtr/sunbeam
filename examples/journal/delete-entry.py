#!/usr/bin/env python3

from journal import load_journal, save_journal

import argparse
parser = argparse.ArgumentParser()
parser.add_argument('--index', type=int, required=True)
args = parser.parse_args()

journal = load_journal()

if args.index < 0 or args.index >= len(journal['entries']):
    raise Exception('Invalid index')

del journal['entries'][args.index]

save_journal(journal)
