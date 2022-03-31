import React from 'react';
import TabbedPageContainer from 'components/TabbedPageContainer';
import DetailsPageWrapper from 'components/DetailsPageWrapper';
import TabDetails from './TabDetails';
import TabResources from './TabResources';

import './package-details.scss';

const DetailsContent = ({data}) => {
    const {id, packageName} = data || {};

    return (
        <TabbedPageContainer
            items={[
                {id: "details", title: "Package details", isIndex: true, component: () => <TabDetails data={data} />},
                {id: "resources", title: "Application resources", path: "resources", component: () => <TabResources id={id} packageName={packageName} /> }
            ]}
        />
    )
}

const PackageDetails = () => (
    <DetailsPageWrapper
        title="Package information"
        backTitle="Packages"
        url="packages"
        detailsContent={DetailsContent}
        getReplace={params => {
            const id = params["id"];
            const innerTab = params["*"];
            
            return !!innerTab ? `/${id}/${innerTab}` : `/${id}`;
        }}
    />
)

export default PackageDetails;