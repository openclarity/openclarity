import React from 'react';
import classnames from 'classnames';

import './label-tag.scss';

export const LabelsDisplay = ({labels, wrapLabels=false}) => (
    <div className={classnames("labels-wrapper", {"wrap-labels": wrapLabels})}>
        {
            labels?.map((label, index) => <div key={index} className="label-tag-wrapper"><LabelTag>{label}</LabelTag></div>)
        }
    </div>
)

const LabelTag = ({children}) => (
    <div className="label-tag">{children}</div>
);

export default LabelTag;