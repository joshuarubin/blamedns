import React, { PropTypes, Component } from 'react'
import { levelToString } from "../lib"

class LogForm extends Component {
    render() {
        let opts = []
        for (let i=5; i > 1; i--) {
            opts.push(i)
        }

        return (
            <form className="log-form row" onSubmit={ev => ev.preventDefault()}>
                <div className="five columns">&nbsp;</div>
                <div className="two columns">
                    <label htmlFor="levelSelector">Log Level</label>
                    <select
                        className="u-full-width"
                        id="levelSelector"
                        value={this.props.logLevel}
                        onChange={this.props.onLevelChange}
                    >
                        {opts.map((value, idx) =>
                            <option key={idx} value={value}>{levelToString(value)}</option>
                        )}
                    </select>
                </div>
            </form>
        );
    }
}

LogForm.propTypes = {
    logLevel: PropTypes.number.isRequired,
    onLevelChange: PropTypes.func.isRequired,
}

export default LogForm
