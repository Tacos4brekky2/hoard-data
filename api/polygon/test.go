package polygon

import (
	// "encoding/json"
	//"context"
	"fmt"
	"io"
	"os"
	// "log"
	//"encoding/csv"
	"encoding/json"
	"net/http"
	"time"
)

// from: unix millisecond time stamp
// Returns daily price data points for a given ticker from a given unix timestamp to now
// If no timestamp given, goes back 2 years (change to 5 with paid api key)
func GetDailyOHLCV(symbol string, fromTime int64) ([]interface{}, error) {
	apiKey := os.Getenv("POLYGON_API_KEY")
	currTime := time.Now().UnixMilli()
	if fromTime == 0 {
		fromTime = currTime - 63072000000 // 2 years in ms
	}
	url := fmt.Sprintf("https://api.polygon.io/v2/aggs/ticker/%s/range/1/day/%d/%d?adjusted=true&sort=asc&apiKey=%s", symbol, fromTime, currTime, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err

	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err

	}

	var response OHLCVResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var result []interface{}
	for _, dataPoint := range response.Results {
		dataPoint.Date = time.Unix(dataPoint.UnixM, 0).UTC()
		dataPoint.Symbol = symbol
		result = append(result, dataPoint)
	}
	return result, nil

}

// APIResponse represents the response from the API.
type OHLCVResponse struct {
	Adjusted     bool          `json:"adjusted" bson:"adjusted"`
	NextURL      string        `json:"next_url" bson:"next_url"`
	QueryCount   int           `json:"queryCount" bson:"query_count"`
	RequestID    string        `json:"request_id" bson:"request_id"`
	Results      []OHLCVWindow `json:"results" bson:"results"`
	ResultsCount int           `json:"resultsCount" bson:"results_count"`
	Status       string        `json:"status" bson:"status"`
	Ticker       string        `json:"ticker" bson:"ticker"`
}

// Result represents each result in the results array.
type OHLCVWindow struct {
	Date           time.Time `bson:"date"`
	UnixM          int64     `json:"t"`
	Symbol         string    `bson:"symbol"`
	Open           float64   `json:"o" bson:"open"`
	High           float64   `json:"h" bson:"high"`
	Low            float64   `json:"l" bson:"low"`
	Close          float64   `json:"c" bson:"close"`
	Volume         float64   `json:"v" bson:"volume"`
	VolumeWeighted float64   `json:"vw" bson:"volume_weighted"`
	NumberOfTrades float64   `json:"n" bson:"number_of_trades"`
}

func GetFinancials(symbol string) (map[string]interface{}, error) {
    apiKey := os.Getenv("POLYGON_API_KEY")
	resp, err := http.Get(fmt.Sprintf("https://api.polygon.io/vX/reference/financials?ticker=%s&timeframe=quarterly&include_sources=true&order=asc&limit=100&sort=filing_date&apiKey=%s", symbol, apiKey))
	if err != nil {
		fmt.Println(err)
		return nil, err

	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err

	}

	var results FinancialResponse
	err = json.Unmarshal(body, &results)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return map[string]interface{}{
		"symbol":     symbol,
		"type":       "financial_statements",
        "last_modified": time.Now(),
		"data": &results.Results,
	}, err
}

// FINANCIAL API RESPONSE DATA
// Financial data structs
type FinancialResponse struct {
	Results   []FinancialResult `json:"results"`
	Status    string            `json:"status"`
	RequestID string            `json:"request_id"`
}

type FinancialStatements struct {
	Symbol     string             `bson:"symbol"`
	Type       string             `bson:"type"`
	Statements *[]FinancialResult `bson:"statements"`
}

type FinancialResult struct {
	StartDate    string        `json:"start_date" bson:"start_date"`
	EndDate      string        `json:"end_date" bson:"end_date"`
	FilingDate   string        `json:"filing_date,omitempty" bson:"filing_date,omitempty"`
	Acceptance   string        `json:"acceptance_datetime,omitempty" bson:"acceptance_datetime,omitempty"`
	Timeframe    string        `json:"timeframe" bson:"timeframe"`
	FiscalPeriod string        `json:"fiscal_period" bson:"fiscal_period"`
	FiscalYear   string        `json:"fiscal_year" bson:"fiscal_year"`
	CIK          string        `json:"cik" bson:"cik"`
	SIC          string        `json:"sic" bson:"sic"`
	Tickers      []string      `json:"tickers" bson:"tickers"`
	CompanyName  string        `json:"company_name" bson:"company_name"`
	SourceURL    string        `json:"source_filing_url,omitempty" bson:"source_filing_url,omitempty"`
	FileURL      string        `json:"source_filing_file_url,omitempty" bson:"source_filing_file_url,omitempty"`
	Financials   FinancialData `json:"financials" bson:"financials"`
}

// FinancialData contains different financial statements.
type FinancialData struct {
	CashFlowStatement   CashFlowStatement   `json:"cash_flow_statement,omitempty" bson:"cash_flow_statement,omitempty"`
	IncomeStatement     IncomeStatement     `json:"income_statement,omitempty" bson:"income_statement,omitempty"`
	BalanceSheet        BalanceSheet        `json:"balance_sheet,omitempty" bson:"balance_sheet,omitempty"`
	ComprehensiveIncome ComprehensiveIncome `json:"comprehensive_income,omitempty" bson:"comprehensive_income,omitempty"`
}

// CashFlowStatement contains cash flow details.
type CashFlowStatement struct {
	NetCashFlow                            ValueUnit `json:"net_cash_flow" bson:"net_cash_flow"`
	NetCashFlowFromOperatingActivities     ValueUnit `json:"net_cash_flow_from_operating_activities" bson:"net_cash_flow_from_operating_activities"`
	NetCashFlowFromOperatingActivitiesCont ValueUnit `json:"net_cash_flow_from_operating_activities_continuing" bson:"net_cash_flow_from_operating_activities_continuing"`
	NetCashFlowFromInvestingActivitiesCont ValueUnit `json:"net_cash_flow_from_investing_activities_continuing" bson:"net_cash_flow_from_investing_activities_continuing"`
	NetCashFlowFromInvestingActivities     ValueUnit `json:"net_cash_flow_from_investing_activities" bson:"net_cash_flow_from_investing_activities"`
	NetCashFlowContinuing                  ValueUnit `json:"net_cash_flow_continuing" bson:"net_cash_flow_continuing"`
	NetCashFlowFromFinancingActivitiesCont ValueUnit `json:"net_cash_flow_from_financing_activities_continuing" bson:"net_cash_flow_from_financing_activities_continuing"`
	NetCashFlowFromFinancingActivities     ValueUnit `json:"net_cash_flow_from_financing_activities" bson:"net_cash_flow_from_financing_activities"`
	ExchangeGainsLosses                    ValueUnit `json:"exchange_gains_losses" bson:"exchange_gains_losses"`
}

// IncomeStatement contains income details.
type IncomeStatement struct {
	NetIncomeLossParent     ValueUnit `json:"net_income_loss_attributable_to_parent" bson:"net_income_loss_attributable_to_parent"`
	IncomeTaxExpenseBenefit ValueUnit `json:"income_tax_expense_benefit" bson:"income_tax_expense_benefit"`
	Revenues                ValueUnit `json:"revenues" bson:"revenues"`
	IncomeLossBeforeTax     ValueUnit `json:"income_loss_from_continuing_operations_before_tax" bson:"income_loss_from_continuing_operations_before_tax"`
	NonOperatingIncomeLoss  ValueUnit `json:"nonoperating_income_loss" bson:"nonoperating_income_loss"`
	GrossProfit             ValueUnit `json:"gross_profit" bson:"gross_profit"`
	OperatingIncomeLoss     ValueUnit `json:"operating_income_loss" bson:"operating_income_loss"`
	ResearchAndDevelopment  ValueUnit `json:"research_and_development" bson:"research_and_development"`
	CostsAndExpenses        ValueUnit `json:"costs_and_expenses" bson:"costs_and_expenses"`
	OperatingExpenses       ValueUnit `json:"operating_expenses" bson:"operating_expenses"`
	BasicEarningsPerShare   ValueUnit `json:"basic_earnings_per_share" bson:"basic_earnings_per_share"`
	DilutedEarningsPerShare ValueUnit `json:"diluted_earnings_per_share" bson:"diluted_earnings_per_share"`
}

// BalanceSheet contains balance sheet details.
type BalanceSheet struct {
	Assets                         ValueUnit `json:"assets" bson:"assets"`
	LiabilitiesAndEquity           ValueUnit `json:"liabilities_and_equity" bson:"liabilities_and_equity"`
	Equity                         ValueUnit `json:"equity" bson:"equity"`
	CurrentAssets                  ValueUnit `json:"current_assets" bson:"current_assets"`
	CurrentLiabilities             ValueUnit `json:"current_liabilities" bson:"current_liabilities"`
	NonCurrentAssets               ValueUnit `json:"noncurrent_assets" bson:"noncurrent_assets"`
	NonCurrentLiabilities          ValueUnit `json:"noncurrent_liabilities" bson:"noncurrent_liabilities"`
	AccountsPayable                ValueUnit `json:"accounts_payable" bson:"accounts_payable"`
	Cash                           ValueUnit `json:"cash" bson:"cash"`
	FixedAssets                    ValueUnit `json:"fixed_assets" bson:"fixed_assets"`
	Wages                          ValueUnit `json:"wages" bson:"wages"`
	OtherCurrentAssets             ValueUnit `json:"other_current_assets" bson:"other_current_assets"`
	OtherCurrentLiabilities        ValueUnit `json:"other_current_liabilities" bson:"other_current_liabilities"`
	OtherNonCurrentAssets          ValueUnit `json:"other_noncurrent_assets" bson:"other_noncurrent_assets"`
	EquityAttributableToParent     ValueUnit `json:"equity_attributable_to_parent" bson:"equity_attributable_to_parent"`
	EquityAttributableToNonControl ValueUnit `json:"equity_attributable_to_noncontrolling_interest" bson:"equity_attributable_to_noncontrolling_interest"`
}

// ComprehensiveIncome contains comprehensive income details.
type ComprehensiveIncome struct {
	ComprehensiveIncomeParent ValueUnit `json:"comprehensive_income_loss_attributable_to_parent" bson:"comprehensive_income_loss_attributable_to_parent"`
	ComprehensiveIncome       ValueUnit `json:"comprehensive_income_loss" bson:"comprehensive_income_loss"`
	OtherComprehensiveIncome  ValueUnit `json:"other_comprehensive_income_loss" bson:"other_comprehensive_income_loss"`
}

// ValueUnit captures the value, unit, and label of financial metrics.
type ValueUnit struct {
	Value float64 `json:"value" bson:"value"`
	Unit  string  `json:"unit" bson:"unit"`
	Label string  `json:"label" bson:"label"`
}
