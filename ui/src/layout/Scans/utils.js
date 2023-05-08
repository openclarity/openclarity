import { isEmpty, isNull } from 'lodash';

export const SCANS_PATHS = {
    SCANS: "scans",
    CONFIGURATIONS: "configurations"
}

export const formatStringInstancesToTags = items => items.map(item => {
    const [key, value] = item.split("=");

    return {key, value};
});

export const formatRegionsToStrings = regions => {
    const SEPARATOR = "/";
    
    return regions?.reduce((acc, curr) => {
        const {name: region, vpcs} = curr;

        const formattedVpcs = vpcs?.reduce((acc, curr) => {
            const {id: vpc, securityGroups} = curr;

            if (!vpc) {
                return acc;
            }

            return [
                ...acc,
                ...(isEmpty(securityGroups) ? [vpc] : securityGroups.map(({id: group}) => `${vpc}${SEPARATOR}${group}`))
            ];
        }, []);
        
        return [
            ...acc,
            ...(isEmpty(formattedVpcs) ? [region] : formattedVpcs.map(formattedVpc => `${region}${SEPARATOR}${formattedVpc}`))
        ]
    }, []);
}

export const getEnabledScanTypesList = scanFamiliesConfig => (
    Object.keys(scanFamiliesConfig).map(scanType => {
        const {enabled} = scanFamiliesConfig[scanType];

        return enabled ? scanType : null;
    })
).filter(scanType => !isNull(scanType));

export const getScanTimeTypeTag = ({operationTime, cronLine}) => {
    if (!!cronLine) {
        return "Repetitive";
    }
    
    if (Date.now() - (new Date(operationTime)).valueOf() <= 0) {
        return "Scheduled";
    }

    return "Once";
}