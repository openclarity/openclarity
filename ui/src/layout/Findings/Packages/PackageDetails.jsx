import React from 'react';
import { useLocation } from 'react-router-dom';
import TabbedPage from 'components/TabbedPage';
import FindingsDetailsPage from '../FindingsDetailsPage';
import TabPackageDetails from './TabPackageDetails';

const PACKAGE_DETAILS_PATHS = {
    PACKAGE_DETAILS: "",
}

const DetailsContent = ({data}) => {
    const {pathname} = useLocation();
    
    const {id} = data;
    
    return (
        <TabbedPage
            basePath={`${pathname.substring(0, pathname.indexOf(id))}${id}`}
            items={[
                {
                    id: "general",
                    title: "Package details",
                    isIndex: true,
                    component: () => <TabPackageDetails data={data} />
                }
            ]}
            withInnerPadding={false}
        />
    )
}

const PackageDetails = () => (
    <FindingsDetailsPage
        backTitle="Packages"
        getTitleData={({findingInfo}) => ({title: findingInfo.name, subTitle: findingInfo.version})}
        detailsContent={DetailsContent}
    />
)

export default PackageDetails;
