#!/usr/bin/env python3

from journal import load_journal, save_journal

import argparse
parser = argparse.ArgumentParser()
parser.add_argument('--uuid', required=True)
args = parser.parse_args()

journal = load_journal()

del journal['entries'][args.uuid]

save_journal(journal)
