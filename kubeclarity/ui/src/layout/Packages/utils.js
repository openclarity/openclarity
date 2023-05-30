import { useFilterDispatch, setFilters, FILTER_TYPES } from 'context/FiltersProvider';
import { OPERATORS } from 'components/Filter';
import InnerAppLink from 'components/InnerAppLink';
import VulnerabilitiesSummaryDisplay from 'components/VulnerabilitiesSummaryDisplay';
import { ApplicationsLink as GeneralApplicationsLink } from 'layout/Applications';
import { ApplicationResourcesLink as GeneralApplicationResourcesLink } from 'layout/ApplicationResources';
import { ROUTES } from 'utils/systemConsts';

const getTitle = name => `package: ${name}`;

export const VulnerabilitiesLink = ({vulnerabilities, id, packageVersion, packageName}) => {
    const filtersDispatch = useFilterDispatch();

    const filterData = [
        {scope: "packageVersion", operator: OPERATORS.is.value, value: [packageVersion]},
        {scope: "packageName", operator: OPERATORS.is.value, value: [packageName]}
    ];

    const onClick = () => {
        setFilters(filtersDispatch, {type: FILTER_TYPES.VULNERABILITIES, filters: filterData, isSystem: false});
    }

    return (
        <InnerAppLink pathname={ROUTES.VULNERABILITIES} onClick={onClick} withUnderline={false}>
            <VulnerabilitiesSummaryDisplay id={id} vulnerabilities={vulnerabilities} withTotal />
        </InnerAppLink>
    )
}

export const ApplicationsLink = ({packageID, applications, packageName}) => (
    <GeneralApplicationsLink count={applications} packageID={packageID} title={getTitle(packageName)} />
)

export const ApplicationResourcesLink = ({packageID, applicationResources, packageName}) => (
    <GeneralApplicationResourcesLink count={applicationResources} packageID={packageID} title={getTitle(packageName)} />
)