import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import { formatDate } from 'utils/utils';


export const getScanColumnsConfigList = () => ([
    {
        Header: "First seen",
        id: "firstSeen",
        sortIds: ["firstSeen"],
        accessor: original => formatDate(original.firstSeen),
    },
    {
        Header: "Last seen",
        id: "lastSeen",
        sortIds: ["lastSeen"],
        accessor: original => formatDate(original.lastSeen)
    }
]);

export const FindingsDetailsCommonFields = ({ firstSeen, lastSeen }) => (
    <>
        <TitleValueDisplayRow>
            <TitleValueDisplay title="First seen">{formatDate(firstSeen)}</TitleValueDisplay>
        </TitleValueDisplayRow>
        <TitleValueDisplayRow>
            <TitleValueDisplay title="Last seen">{formatDate(lastSeen)}</TitleValueDisplay>
        </TitleValueDisplayRow>
    </>
)
