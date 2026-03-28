export type IInvoiceStatus = "NEW" | "CONCEPT" | "FINAL";
export interface IInvoiceProps {
  id?: string;
  entity: string;
  year: string;
  bucket?: string;
}
export interface IInvoiceEntity {
  Name: string
  Street1: string
  Street2: string
}
export interface IInvoiceCustomer {
  Name: string
  Street1: string
  Street2: string
  Vat: string
  Coc: string
  Tax: string
}
export interface IInvoiceMeta {
  Conceptid: string;
  Status: IInvoiceStatus;
  Invoiceid: string;
  Issuedate?: string | null;
  Ponumber: string;
  Duedate?: string | null;
  Paydate?: string | null;
  Freefield?: string;
  HourFile: string;
}
export interface IInvoiceLine {
  Description: string
  Quantity: string //number
  Price: string //number
  Total: string //number
}
export interface IInvoiceTotal {
  Ex: string //number
  Tax: string //number
  Total: string //number
}
export interface IInvoiceBank {
  Vat: string;
  Coc: string;
  Iban: string;
  Bic: string;
}
export interface IInvoiceMail {
  From: string
  Subject: string
  To: string
  Body: string
}
export interface IInvoiceState {
  Company?: string
  Entity?: IInvoiceEntity
  Customer?: IInvoiceCustomer
  Meta?: IInvoiceMeta
  Lines?: IInvoiceLine[]
  Notes?: string
  Total?: IInvoiceTotal
  Bank?: IInvoiceBank
  Mail?: IInvoiceMail
}