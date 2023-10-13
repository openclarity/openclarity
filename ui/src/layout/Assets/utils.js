export function getAssetName(assetInfo) {
    switch (assetInfo.objectType) {
        case "VMInfo":
            return assetInfo.instanceID;
        case "PodInfo":
            return assetInfo.podName;
        case "DirInfo":
            return assetInfo.dirName;
        case "ContainerImageInfo":
            return assetInfo.name;
        case "ContainerInfo":
            return assetInfo.containerName;
        default:
            return assetInfo.id;
    }
}
