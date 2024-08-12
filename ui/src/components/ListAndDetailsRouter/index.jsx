import React from 'react';
import { Route, Routes, Outlet } from 'react-router-dom';

const ListAndDetailsRouter = ({listComponent: ListComponent, detailsComponent: DetailsComponent, detailsPath=":id"}) => (
    <Routes>
        <Route path="/" element={<Outlet />}>
            <Route index element={<ListComponent />} />
            <Route path={`${detailsPath}/*`} element={<DetailsComponent />} />
        </Route>
    </Routes>
)

export default ListAndDetailsRouter;