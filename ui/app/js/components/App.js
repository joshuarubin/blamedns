import React from 'react';
import { Link } from 'react-router'

const App = () => (
    <ul className="nav nav-pills nav-stacked">
        <li role="presentation">
            <Link to="/ui/logs">logs</Link>
        </li>
    </ul>
)

export default App
