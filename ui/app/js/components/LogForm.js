import React, { PropTypes, Component } from 'react'
import { levelToString } from "../lib"

class LogForm extends Component {
    render() {
        let opts = []
        for (let i=5; i > 1; i--) {
            opts.push({
                level: i,
                name: levelToString(i),
                active: i === this.props.logLevel
            })
        }

        return (
            <nav className="navbar navbar-fixed-top navbar-default">
                <div className="container-fluid">
                    <span className="navbar-brand">blamedns</span>
                    <div className="navbar-right">
                        <p className="navbar-text">Log Level</p>
                        <div className="navbar-form btn-group" data-toggle="buttons">
                            {opts.map((value, idx) =>
                                <label key={idx} className={"btn btn-default"+(value.active ? " active" : "")}>
                                    <input
                                        type="radio"
                                        name="log-level"
                                        autocomplete="off"
                                        value={value.level}
                                        defaultChecked={value.active}
                                        onChange={this.props.onLevelChange}
                                    />
                                    {value.name}
                                </label>
                            )}
                        </div>
                    </div>
                </div>
            </nav>
        );
    }
}

LogForm.propTypes = {
    logLevel: PropTypes.number.isRequired,
    onLevelChange: PropTypes.func.isRequired,
}

export default LogForm
