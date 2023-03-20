import React, { useEffect } from 'react';
import { usePrevious, useFetch } from 'hooks';
import { TextField, RadioField, MultiselectField, SelectField, FieldsPair, useFormikContext, CheckboxField,
    FieldLabel, validators } from 'components/Form';
import { APIS } from 'utils/systemConsts';

export const VPCS_EMPTY_VALUE = [{id: "", securityGroups: []}]
export const REGIONS_EMPTY_VALUE = [{name: "", vpcs: VPCS_EMPTY_VALUE}];

export const SCOPE_ITEMS = {
    ALL: {value: "ALL", label: "All"},
    DEFINED: {value: "DEFINED", label: "Define scope"}
}

const SecurityGroupField = ({index, name, placeholder, disabled, regionData}) => {
    const {name: regionId, vpcs=[]} = regionData || {};
    const vpcId = vpcs[index]?.id;
    const prevVpsId = usePrevious(vpcId);
    
    const [{data, loading, error}, fetchSecurityGroups] = useFetch(APIS.SCOPES_DISCOVERY, {loadOnMount: false});

    useEffect(() => {
        if (!!vpcId && prevVpsId !== vpcId) {
            fetchSecurityGroups({queryParams: {
                "$select": `scopeInfo/regions($filter=name eq '${regionId}';$select=name,vpcs($filter=id eq '${vpcId}';$select=securityGroups))`
            }});
        }
    }, [prevVpsId, vpcId, regionId, fetchSecurityGroups]);
    
    if (error) {
        return null;
    }

    return (
        <MultiselectField
            name={name}
            placeholder={placeholder}
            disabled={disabled}
            items={!data ? [] : data?.scopeInfo?.regions[0]?.vpcs[0].securityGroups.map(({id}) => ({value: id, label: id}))}
            loading={loading}
        />
    )
}

const RegionFields = ({index, name, disabled}) => {
    const {values} = useFormikContext();
    const {regions} = values.scope;
    
    const regionId = regions[index]?.name;
    const prevRegionId = usePrevious(regionId);

    const [{data, loading, error}, fetchVpcs] = useFetch(APIS.SCOPES_DISCOVERY, {loadOnMount: false});

    useEffect(() => {
        if (!!regionId && prevRegionId !== regionId) {
            fetchVpcs({queryParams: {
                "$select": `scopeInfo/regions($filter=name eq '${regionId}';$select=id,vpcs/id)`,
            }});
        }
    }, [prevRegionId, regionId, fetchVpcs]);
    
    if (error) {
        return null;
    }
    
    return (
        <FieldsPair
            name={name}
            disabled={disabled}
            firstFieldProps={{
                component: SelectField,
                key: "id",
                placeholder: "Select VPC...",
                items: !data ? [] : data?.scopeInfo?.regions[0]?.vpcs.map(({id}) => ({value: id, label: id})),
                clearable: true,
                loading: loading
            }}
            secondFieldProps={{
                component: SecurityGroupField,
                key: "securityGroups",
                placeholder: "Select security group...",
                emptyValue: [],
                regionData: regions[index]
            }}
        />
    );
}

const DefinedScopeFields = () => {
    const [{data, loading, error}] = useFetch(APIS.SCOPES_DISCOVERY, {queryParams: {"$select": "scopeInfo/regions/name"}});

    if (error) {
        return null;
    }

    return (
        <FieldsPair
            name="scope.regions"
            firstFieldProps={{
                component: SelectField,
                key: "name",
                placeholder: "Select region...",
                items: data?.scopeInfo?.regions.map(({name}) => ({value: name, label: name})),
                validate: validators.validateRequired,
                loading: loading
            }}
            secondFieldProps={{
                component: RegionFields,
                key: "vpcs",
                emptyValue: VPCS_EMPTY_VALUE
            }}
        />
    )
}

const StepGeneralProperties = () => {
    const {values, setFieldValue} = useFormikContext();
    const {scopeSelect} = values.scope;

    const isDefineScope = scopeSelect === SCOPE_ITEMS.DEFINED.value;
    const prevIsDefineScope = usePrevious(isDefineScope);
    
    useEffect(() => {
        if (prevIsDefineScope && !isDefineScope) {
            setFieldValue("scope.regions", REGIONS_EMPTY_VALUE);
        }
    }, [prevIsDefineScope, isDefineScope, setFieldValue]);
    
    return (
        <div className="scan-config-general-step">
            <TextField
                name="name"
                label="Scan config name*"
                placeholder="Type a name..."
                validate={validators.validateRequired}
            />
            <RadioField
                name="scope.scopeSelect"
                label="Scope"
                items={Object.values(SCOPE_ITEMS)}
            />
            {isDefineScope && <DefinedScopeFields />}
            <div className="scan-config-instances">
                <FieldLabel>Instances</FieldLabel>
                <CheckboxField
                    name="scope.shouldScanStoppedInstances"
                    title="Scan also non-running instances"
                />
                <MultiselectField
                    name="scope.instanceTagSelector"
                    placeholder="Select instances to include (key=value)..."
                    connector="and"
                    prefixLabel="Include"
                    creatable
                    validate={validators.keyValueListValidator}
                />
                <MultiselectField
                    name="scope.instanceTagExclusion"
                    placeholder="Select instances to exclude (key=value)..."
                    connector="and"
                    prefixLabel="Exclude"
                    creatable
                    validate={validators.keyValueListValidator}
                />
            </div>
        </div>
    )
}

export default StepGeneralProperties;