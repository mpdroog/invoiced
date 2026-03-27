import * as React from "react";
import Axios from "axios";
import "./search.css";

interface SearchResult {
    type: string;
    id: string;
    title: string;
    subtitle: string;
    entity: string;
    year: number;
    quarter: number;
    bucket: string;
}

interface MenuState {
    ahead: number;
    searchQuery: string;
    searchResults: SearchResult[];
    showResults: boolean;
    searching: boolean;
}

class Menu extends React.Component<{company: string, year: string}, MenuState> {
    private boundRefresh: () => void;
    private searchTimeout: any = null;

    constructor(props) {
        super(props);
        this.state = {
            ahead: 0,
            searchQuery: '',
            searchResults: [],
            showResults: false,
            searching: false
        };
        this.boundRefresh = this.fetchGitStatus.bind(this);
    }

    componentDidMount() {
        this.fetchGitStatus();
        window.addEventListener('git-refresh', this.boundRefresh);
        document.addEventListener('click', this.handleClickOutside.bind(this));
    }

    componentWillUnmount() {
        window.removeEventListener('git-refresh', this.boundRefresh);
        document.removeEventListener('click', this.handleClickOutside.bind(this));
    }

    private handleClickOutside(e: MouseEvent) {
        const searchContainer = document.getElementById('search-container');
        if (searchContainer && !searchContainer.contains(e.target as Node)) {
            this.setState({showResults: false});
        }
    }

    private fetchGitStatus() {
        Axios.get('/api/v1/git/' + this.props.company + '/status')
            .then(res => {
                this.setState({ahead: res.data.ahead || 0});
            })
            .catch(() => {
                // Silently ignore - git status may not be available
            });
    }

    private handleSearchChange(e: React.ChangeEvent<HTMLInputElement>) {
        const query = e.target.value;
        this.setState({searchQuery: query, showResults: true});

        if (this.searchTimeout) {
            clearTimeout(this.searchTimeout);
        }

        if (query.length < 2) {
            this.setState({searchResults: [], searching: false});
            return;
        }

        this.setState({searching: true});
        this.searchTimeout = setTimeout(() => {
            Axios.get('/api/v1/search/' + this.props.company, {params: {q: query}})
                .then(res => {
                    this.setState({searchResults: res.data.results || [], searching: false});
                })
                .catch(() => {
                    this.setState({searchResults: [], searching: false});
                });
        }, 300);
    }

    private handleResultClick(result: SearchResult) {
        let url = '#' + result.entity + '/' + result.year + '/';
        if (result.type === 'invoice') {
            url += 'invoices/edit/' + result.bucket + '/' + result.id;
        } else if (result.type === 'hour') {
            url += 'hours/edit/' + result.bucket + '/' + result.id;
        } else if (result.type === 'purchase') {
            url += 'purchases';
        }
        this.setState({showResults: false, searchQuery: ''});
        location.hash = url;
    }

    private getResultIcon(type: string): string {
        switch (type) {
            case 'invoice': return 'fa-money';
            case 'hour': return 'fa-clock-o';
            case 'purchase': return 'fa-shopping-cart';
            default: return 'fa-file';
        }
    }

    render() {
        const {company, year} = this.props;
        const {ahead, searchQuery, searchResults, showResults, searching} = this.state;

        return (<div id="header">
            <div id="logo" className="light-version">
                <a href="/" id="js-entities"><img src={"/api/v1/entities/" + company + "/logo"} className="m-b" alt="logo"/></a>
            </div>
            <nav role="navigation">
                <div className="nav-spacer">&nbsp;</div>

                <div id="search-container" className="navbar-form-custom search-container">
                    <div className="form-group">
                        <input
                            type="text"
                            placeholder="Search invoices, hours..."
                            className="form-control"
                            id="js-search"
                            value={searchQuery}
                            onChange={this.handleSearchChange.bind(this)}
                            onFocus={() => this.setState({showResults: true})}
                        />
                    </div>
                    {showResults && searchQuery.length >= 2 && (
                        <div className="search-dropdown">
                            {searching ? (
                                <div className="search-loading">
                                    <i className="fa fa-spinner fa-spin"></i> Searching...
                                </div>
                            ) : searchResults.length > 0 ? (
                                searchResults.map((r, i) => (
                                    <div
                                        key={i}
                                        className="search-result-item"
                                        onClick={() => this.handleResultClick(r)}
                                    >
                                        <i className={'fa ' + this.getResultIcon(r.type) + ' search-result-icon'}></i>
                                        <strong>{r.title}</strong>
                                        <span className="search-result-subtitle">{r.subtitle}</span>
                                        <div className="search-result-meta">
                                            {r.type} &middot; {r.year} Q{r.quarter}
                                        </div>
                                    </div>
                                ))
                            ) : (
                                <div className="search-no-results">
                                    No results found
                                </div>
                            )}
                        </div>
                    )}
                </div>
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
                        <li>
                            <a href={"#"+company+"/"+year+"/purchases"} id="js-purchases" title="Purchase Invoices">
                                <i className="fa fa-shopping-cart"></i>
                            </a>
                        </li>
                        <li>
                            <a href={"#"+company+"/"+year+"/taxes"} id="js-taxes">
                                <i className="fa fa-bank"></i>
                            </a>
                        </li>
                        <li>
                            <a href={"#"+company+"/"+year+"/git"} id="js-git">
                                <i className="fa fa-cloud-upload"></i>
                                {ahead > 0 && <span className="label label-danger">{ahead}</span>}
                            </a>
                        </li>
                    </ul>
                </div>
            </nav>
          </div>
        );
    }
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
