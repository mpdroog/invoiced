import * as React from "react";

function Menu ({ company, year }) {
    var logo = "/static/assets/"+company+".png";
    var s = {marginRight: "15px", float: "left"};
    return (<div id="header">
        <div id="logo" className="light-version">
            <a href="/" id="js-entities"><img src={logo} className="m-b" alt="logo"/></a>
        </div>
        <nav role="navigation">
            <div style={s}>&nbsp;</div>

            <form role="search" className="navbar-form-custom" method="post" action="#">
                <div className="form-group">
                    <input type="text" placeholder="Search something special" className="form-control" name="search" id="js-search"/>
                </div>
            </form>
            <div className="navbar-right">
                <ul className="nav navbar-nav no-borders">
                    <li>
                        <a href={"#"+company+"/"+year} id="js-dashboard">
                            <i className="fa fa-dashboard"></i>
                        </a>
                    </li>
                    <li>
                        <a href={"#"+company+"/"+year+"/hours"} id="js-hours">
                            <i className="fa fa-clock-o"></i>
                        </a>
                    </li>
                    <li>
                        <a href={"#"+company+"/"+year+"/invoices"} id="js-invoices">
                            <i className="fa fa-money"></i>
                        </a>
                    </li>
                </ul>
            </div>
        </nav>
    </div>);
}

export function calendar({ year }) {
    let selected = {fontSize:"18px"};
    return (<div>
        <div className="month"> 
          <ul>
            <li className="prev">&#10094;</li>
            <li className="next">&#10095;</li>
            <li>
              August<br/>
              <span style={selected}>{year}</span>
            </li>
          </ul>
        </div>

        <ul className="weekdays">
          <li>Mo</li>
          <li>Tu</li>
          <li>We</li>
          <li>Th</li>
          <li>Fr</li>
          <li>Sa</li>
          <li>Su</li>
        </ul>

        <ul className="days"> 
          <li>1</li>
          <li>2</li>
          <li>3</li>
          <li>4</li>
          <li>5</li>
          <li>6</li>
          <li>7</li>
          <li>8</li>
          <li>9</li>
          <li><span className="active">10</span></li>
          <li>11</li>
        </ul>
    </div>);
}

export class Design extends React.Component {
  constructor(props) {
    super(props);
  }
  render() {
    return <div>
        <Menu company={this.props.entity} year={this.props.year}/>
        {this.props.children}
    </div>;
  }
}
