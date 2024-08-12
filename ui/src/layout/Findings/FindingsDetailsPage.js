import React from 'react';
import DetailsPageWrapper from 'components/DetailsPageWrapper';
import { APIS } from 'utils/systemConsts';

const FindingsDetailsPage = ({backTitle, getTitleData, detailsContent: DetailsContent}) => (
    <DetailsPageWrapper
        backTitle={backTitle}
        url={APIS.FINDINGS}
        select="id,scan,asset,findingInfo,foundOn,invalidatedOn"
        expand="asset($select=id,assetInfo),scan($select=id,scanConfig,scanConfigSnapshot,startTime,endTime,summary,state,stateMessage,stateReason)"
        getTitleData={getTitleData}
        detailsContent={props => <DetailsContent {...props} />}
    />
)

export default FindingsDetailsPage;
