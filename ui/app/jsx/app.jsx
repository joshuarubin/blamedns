import React from 'react';
import { Link } from 'react-router'

export default React.createClass({
    render() {
        return (
            <div>
                <ul>
                    <li><Link to="/ui/logs">logs</Link></li>
                </ul>
            </div>
        );
    }
});
