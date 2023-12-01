import React from 'react';
import copy from 'copy-to-clipboard';
import Button from 'components/Button';

const CopyButton = ({copyText}) => (
    <Button tertiary onClick={() => copy(copyText, {format: "text/plain"})} >
        Copy
    </Button>
)

export default CopyButton;