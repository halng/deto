#!/bin/bash

# Fetch the latest Java versions (Adoptium)
curl -s https://api.adoptium.net/v3/info/release_versions | jq '.[] | .version' > java_versions.json

# Fetch the latest Go versions
curl -s https://go.dev/dl/?json=1 | jq '.[] | select(.stable==true) | .version' > go_versions.json

# Fetch the latest Python versions (Python.org)
curl -s https://www.python.org/doc/versions/ | grep -oP '(?<=<span class="version">)\d+\.\d+\.\d+' > python_versions.txt

# Fetch the latest Node.js versions
curl -s https://nodejs.org/dist/index.json | jq '.[] | select(.lts != null) | .version' > node_versions.json

# Combine all the versions into one JSON file
echo "{" > registry.json
echo "\"java\": $(cat java_versions.json)," >> registry.json
echo "\"go\": $(cat go_versions.json)," >> registry.json
echo "\"python\": $(cat python_versions.txt | jq -R . | jq -s .)," >> registry.json
echo "\"nodejs\": $(cat node_versions.json)" >> registry.json
echo "}" >> registry.json

# Check if the registry.json file has changed
if ! git diff --quiet registry.json; then
  # If changes are found, add and commit them
  git config --local user.email "your-email@example.com"
  git config --local user.name "Your Name"
  git add registry.json
  git commit -m "Updated registry with latest versions"

  # Push the changes to a new branch
  git checkout -b update-versions-$(date +%Y%m%d%H%M%S)
  git push --set-upstream origin update-versions-$(date +%Y%m%d%H%M%S)

  # Create a pull request using GitHub CLI (ensure the GitHub CLI is installed)
  gh pr create --title "Update Java, Go, Python, Node.js versions" --body "This PR updates the registry with the latest versions." --base main --head update-versions-$(date +%Y%m%d%H%M%S)
else
  echo "No changes in registry.json, skipping commit and PR creation."
fi
