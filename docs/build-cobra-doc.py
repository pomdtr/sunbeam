import subprocess
import sys
import pathlib
import os

output_dir = pathlib.Path(__file__).parent / "src" / "cli"
try:
    env = os.environ.copy()
    env["DISABLE_EXTENSIONS"] = "1"
    subprocess.run(["sunbeam", "generate-docs", str(output_dir.absolute())], env=env, text=True, check=True)
except subprocess.CalledProcessError as e:
    print("Error running sunbeam generate-docs", file=sys.stderr)
    print(e, file=sys.stderr)
    sys.exit(1)


toc = []
for f in sorted(output_dir.iterdir(), key=lambda x: x.name):
    title = f.stem.replace("_", " ")
    indentation_level = len(title.split(" ")) - 1
    toc.append(f"{'  ' * indentation_level}- [{title}](./cli/{f.name})")

input_summary = pathlib.Path(__file__).parent / "src" / "_SUMMARY.md"
with open(input_summary, "r") as f:
    summary = f.read()

summary = summary.replace("<!-- CLI TOC -->", "\n".join(toc))

output_summary = pathlib.Path(__file__).parent / "src" / "SUMMARY.md"
with open(output_summary, "w") as f:
    f.write(summary)
