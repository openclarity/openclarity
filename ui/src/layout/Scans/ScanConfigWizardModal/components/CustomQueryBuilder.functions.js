const convertDataType = (originalType) => {
    switch (originalType) {
        case "string":
            return "text";
        case "integer":
            return "number";
        case "object":
            return "!struct";
        case "array":
            return "!struct"
        default:
            return originalType;
    }
}

const postConvertQuery = (query) => {
    let result = query?.replace(/&&/g, "and");
    return result;
}

export { convertDataType, postConvertQuery };
