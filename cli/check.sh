#!/bin/bash

failed=0

echo "Running ShellCheck on CLI shell scripts..."
[ -f "task" ] && { echo "Checking: task"; shellcheck -s sh "task" || failed=1; }
for f in $(find lib -name "*.sh" -type f 2>/dev/null); do
    [ -n "$f" ] && { echo "Checking: $f"; shellcheck -s sh "$f" || failed=1; }
done
for f in $(find test -name "*.sh" -type f 2>/dev/null); do
    [ -n "$f" ] && { echo "Checking: $f"; shellcheck -s sh "$f" || failed=1; }
done
if [ "$failed" -eq 1 ]; then
    echo "ShellCheck found errors!"
    exit 1
fi
echo "ShellCheck completed successfully."
