import requests
import json

# ============================================ Java Versions ============================================
# This script fetches all available Java versions from different providers and saves them to a JSON file
# Supported providers: Adoptium
# =======================================================================================================


def _adoptium():
    """ 
    Fetches all available Java versions from Adoptium and saves them to a JSON file
    Only includes versions from the available_releases list to ensure validity
    Keeps only the 2 latest versions for each major version to reduce duplicates
    Docs: https://api.adoptium.net/q/swagger-ui/
    """
    GET_ALL_VERSIONS_API = "https://api.adoptium.net/v3/info/available_releases"
    GET_VERSION_DETAILS_API = "https://api.adoptium.net/v3/assets/feature_releases/{version}/ga?image_type=jdk"

    req = requests.get(GET_ALL_VERSIONS_API)
    versions = req.json()

    available_versions = versions["available_releases"]
    all_version = {}
    # Dictionary to track versions by major version and OS/arch combination
    version_tracking = {}

    for version in available_versions:
        # Validate version is in the available releases
        if version not in available_versions:
            print(f"Skipping invalid version: {version}")
            continue
            
        print(f"Fetching details for Java {version} from Adoptium...")
        req = requests.get(GET_VERSION_DETAILS_API.format(version=version))
        
        # Validate API response
        if req.status_code != 200:
            print(f"Failed to fetch details for Java {version}, status code: {req.status_code}")
            continue
            
        details = req.json()
        for obj in details:
            binaries = obj["binaries"]
            for binary in binaries:
                # Validate required fields exist
                if not all(k in binary for k in ['os', 'architecture', 'package']):
                    print(f"Skipping binary with missing fields for Java {version}")
                    continue
                
                os_name = binary['os']
                arch = binary["architecture"]
                
                # Track versions for deduplication
                key = (version, os_name, arch)
                package_name = binary["package"]["name"]
                
                if key not in version_tracking:
                    version_tracking[key] = []
                
                version_tracking[key].append({
                    "architecture": arch,
                    "checksum": binary["package"]["checksum"],
                    "link": binary["package"]["link"],
                    "name": package_name,
                    "version": str(version),
                    "is_lts": version in versions["available_lts_releases"],
                    "provider": "Adoptium",
                    "update_version": obj.get("version_data", {}).get("openjdk_version", package_name)
                })

    # Now filter to keep only the 2 latest versions for each major version/os/arch combo
    for (major_version, os_name, arch), entries in version_tracking.items():
        # Sort by update version (extract the patch/build number from the name)
        # Names follow pattern like: OpenJDK8U-jdk_x64_linux_hotspot_8u472b08.tar.gz
        # We want to sort by the version number (e.g., 8u472b08)
        sorted_entries = sorted(entries, key=lambda x: x['name'], reverse=True)
        
        # Keep only the 2 latest
        latest_entries = sorted_entries[:2]
        
        for entry in latest_entries:
            if os_name not in all_version:
                all_version[os_name] = []
            
            # Remove the temporary update_version field
            entry_to_add = {k: v for k, v in entry.items() if k != 'update_version'}
            all_version[os_name].append(entry_to_add)

    with open("../registry/java_versions.json", "w") as f:
        json.dump(all_version, f, indent=2)


def fetch_all_java_versions():
    """
    Fetches all available Java versions from different providers and saves them to a JSON file
    """
    _adoptium()
    # add other providers here

# ============================================ Go Versions ============================================
# This script fetches all available Go versions from source and saves them to a JSON file
# =======================================================================================================


def fetch_all_go_versions():
    """
    Fetches all available Go versions from source and saves them to a JSON file
    Includes all versions and tracks stability with is_stable field
    """
    GET_ALL_VERSIONS_API = "https://go.dev/dl"
    response = requests.get(f"{GET_ALL_VERSIONS_API}/?mode=json&include=all")
    versions = response.json()
    all_versions = {}
    
    # Process all versions (stable and unstable)
    for ver in versions:
        is_stable = ver.get('stable', False)
            
        for file in ver['files']:
          if file['os'] != '' and file['arch'] != '' and file['kind'] == 'archive':
            if file['os'] not in all_versions:
              all_versions[file['os']] = []
            
            all_versions[file['os']].append({
                "version": str(file['version']),
                "architecture": file['arch'],
                "name": file['filename'],
                "checksum": file['sha256'],
                "provider": "Open Source",
                "is_stable": is_stable,
                "link": f"{GET_ALL_VERSIONS_API}/{file['filename']}"
            })
    
    with open("../registry/go_versions.json", "w") as f:
        json.dump(all_versions, f, indent=2)


# ======================== Main ========================
if __name__ == "__main__":
    fetch_all_java_versions()
    fetch_all_go_versions()
