import React from 'react';
import DetailsPageWrapper from 'components/DetailsPageWrapper';
import { APIS } from 'utils/systemConsts';

const FindingsDetailsPage = ({backTitle, getTitleData, detailsContent: DetailsContent}) => (
    <DetailsPageWrapper
        backTitle={backTitle}
        url={APIS.FINDINGS}
        select="id,findingInfo,foundOn,invalidatedOn"
        expand="asset($select=id,assetInfo,firstSeen,lastSeen,terminatedOn)"
        getTitleData={getTitleData}
        detailsContent={props => <DetailsContent {...props} />}
    />
)

export default FindingsDetailsPage;
