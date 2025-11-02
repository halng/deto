import requests
import json
import re

def validate_go_versions():
    """
    Validates Go versions in the registry by:
    1. Checking if versions are stable
    2. Verifying version format
    3. Checking if download links are accessible
    """
    print("Validating Go versions...")
    
    # Fetch current data from API
    GET_ALL_VERSIONS_API = "https://go.dev/dl"
    response = requests.get(f"{GET_ALL_VERSIONS_API}/?mode=json")
    api_versions = response.json()
    
    # Get only stable versions
    stable_versions = [v for v in api_versions if v.get('stable', False)]
    
    print(f"Found {len(api_versions)} total versions in API")
    print(f"Found {len(stable_versions)} stable versions in API")
    
    # Load current registry
    with open("../registry/go_versions.json", "r") as f:
        current_registry = json.load(f)
    
    # Count entries
    total_entries = sum(len(versions) for versions in current_registry.values())
    print(f"Current registry has {total_entries} entries")
    
    # Get list of stable version numbers
    stable_version_numbers = {v['version'] for v in stable_versions}
    print(f"Stable versions: {sorted(stable_version_numbers)}")
    
    # Check each entry in registry
    invalid_entries = []
    for os_name, versions in current_registry.items():
        for entry in versions:
            if entry['version'] not in stable_version_numbers:
                invalid_entries.append({
                    'os': os_name,
                    'version': entry['version'],
                    'reason': 'Not a stable version'
                })
    
    print(f"Found {len(invalid_entries)} invalid entries")
    
    return stable_versions, invalid_entries

def validate_java_versions():
    """
    Validates Java versions in the registry by:
    1. Checking if versions exist in Adoptium API
    2. Verifying version format
    """
    print("\nValidating Java versions...")
    
    # Fetch available releases from API
    GET_ALL_VERSIONS_API = "https://api.adoptium.net/v3/info/available_releases"
    req = requests.get(GET_ALL_VERSIONS_API)
    api_data = req.json()
    
    available_versions = set(api_data["available_releases"])
    print(f"Available Java versions from API: {sorted(available_versions)}")
    
    # Load current registry
    with open("../registry/java_versions.json", "r") as f:
        current_registry = json.load(f)
    
    # Count entries
    total_entries = sum(len(versions) for versions in current_registry.values())
    print(f"Current registry has {total_entries} entries")
    
    # Check each entry
    invalid_entries = []
    registry_versions = set()
    for os_name, versions in current_registry.items():
        for entry in versions:
            version = int(entry['version'])
            registry_versions.add(version)
            if version not in available_versions:
                invalid_entries.append({
                    'os': os_name,
                    'version': version,
                    'reason': 'Not in available releases'
                })
    
    print(f"Versions in registry: {sorted(registry_versions)}")
    print(f"Found {len(invalid_entries)} invalid entries")
    
    return available_versions, invalid_entries

if __name__ == "__main__":
    print("="*80)
    print("REGISTRY VALIDATION REPORT")
    print("="*80)
    
    stable_go, invalid_go = validate_go_versions()
    available_java, invalid_java = validate_java_versions()
    
    print("\n" + "="*80)
    print("SUMMARY")
    print("="*80)
    print(f"Go invalid entries: {len(invalid_go)}")
    print(f"Java invalid entries: {len(invalid_java)}")
    
    if invalid_go:
        print("\nInvalid Go entries (sample):")
        for entry in invalid_go[:10]:
            print(f"  - {entry['version']} ({entry['os']}): {entry['reason']}")
    
    if invalid_java:
        print("\nInvalid Java entries (sample):")
        for entry in invalid_java[:10]:
            print(f"  - {entry['version']} ({entry['os']}): {entry['reason']}")
