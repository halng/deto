import requests
import json

# ============================================ Java Versions ============================================
# This script fetches all available Java versions from different providers and saves them to a JSON file
# Supported providers: Adoptium
# =======================================================================================================


def _adoptium():
    """ 
    Fetches all available Java versions from Adoptium and saves them to a JSON file
    Docs: https://api.adoptium.net/q/swagger-ui/
    """
    GET_ALL_VERSIONS_API = "https://api.adoptium.net/v3/info/available_releases"
    GET_VERSION_DETAILS_API = "https://api.adoptium.net/v3/assets/feature_releases/{version}/ga?image_type=jdk"

    req = requests.get(GET_ALL_VERSIONS_API)
    versions = req.json()

    available_versions = versions["available_releases"]
    all_version = {}

    for version in available_versions:
        print(f"Fetching details for Java {version} from Adoptium...")
        req = requests.get(GET_VERSION_DETAILS_API.format(version=version))
        details = req.json()
        for obj in details:
            binaries = obj["binaries"]
            for binary in binaries:
                if binary['os'] not in all_version:
                    all_version[binary['os']] = []
                all_version[binary['os']].append({
                    "architecture": binary["architecture"],
                    "checksum": binary["package"]["checksum"],
                    "link": binary["package"]["link"],
                    "name": binary["package"]["name"],
                    "version": str(version),
                    "is_lts": version in versions["available_lts_releases"],
                    "provider": "Adoptium"
                })

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
    """
    GET_ALL_VERSIONS_API = "https://go.dev/dl"
    response = requests.get(f"{GET_ALL_VERSIONS_API}/?mode=json")
    versions = response.json()
    all_versions = {}
    for ver in versions:
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
                "is_lts": True,
                "link": f"{GET_ALL_VERSIONS_API}/{file['filename']}"
            })
    
    with open("../registry/go_versions.json", "w") as f:
        json.dump(all_versions, f, indent=2)


# ======================== Main ========================
if __name__ == "__main__":
    fetch_all_java_versions()
    fetch_all_go_versions()
