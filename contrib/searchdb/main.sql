/*
 Source Server Type    : SQLite
 Source Server Version : 3007015
 Source Database       : main

 Target Server Type    : SQLite
 Target Server Version : 3007015
 File Encoding         : utf-8

 Date: 06/07/2018 22:02:04 PM
*/

PRAGMA foreign_keys = false;

-- ----------------------------
--  Table structure for "invoice_lines"
-- ----------------------------
DROP TABLE IF EXISTS "invoice_lines";
CREATE TABLE "invoice_lines" (
	 "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	 "invoices_id" integer NOT NULL,
	 "description" text NOT NULL,
	 "quantity" integer NOT NULL,
	 "price" real NOT NULL,
	 "total" real NOT NULL
);

-- ----------------------------
--  Table structure for "invoices"
-- ----------------------------
DROP TABLE IF EXISTS "invoices";
CREATE TABLE "invoices" (
	 "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	 "total_ex" real NOT NULL,
	 "total_tax" real NOT NULL,
	 "total_sum" real NOT NULL,
	 "concept_id" text NOT NULL,
	 "invoice_id" text NOT NULL,
	 "customer_name" text NOT NULL
);
CREATE UNIQUE INDEX "main"."unique_invoice" ON "invoices" ("invoice_id" ASC);

PRAGMA foreign_keys = true;
