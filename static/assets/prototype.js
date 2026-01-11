// Small bunch of prototype functions to rapidly
// play with data in memory
//
// Using defineProperty to make methods non-enumerable (React compatible)

Object.defineProperty(Object.prototype, 'concatArray', {
	enumerable: false,
	value: function() {
		var n = Object.values(this);
		var out = [];
		for (var i = 0; i < n.length; i++) {
			out = out.concat(n[i]);
		}
		return out;
	}
});

Object.defineProperty(Array.prototype, 'filterField', {
	enumerable: false,
	value: function(fieldname) {
		var out = [];
		var keys = fieldname.split(".");
		for (var i = 0; i < this.length; i++) {
			var base = this[i];
			for (var n = 0; n < keys.length; n++) {
				base = base[ keys[n] ];
			}
			out.push(base);
		}
		return out;
	}
});

Object.defineProperty(Array.prototype, 'remove', {
	enumerable: false,
	value: function(id) {
		// ignore res
		this.splice(id, 1);
		return this;
	}
});

Object.defineProperty(Array.prototype, 'sum', {
	enumerable: false,
	value: function(round) {
		var out = new Big("0.00");

		for (var i = 0; i < this.length; i++) {
			out = out.plus(new Big(this[i]));
		}

		return out.round(round).toFixed(round).toString();
	}
});

console.log(`
Total ex VAT this year:
window.rootdev.invoiced.paid.concatArray().filterField("Total.Ex").sum(2);
Get worked hours this year:
window.rootdev.invoiced.paid.concatArray().filterField("Lines").concatArray().filterField("Quantity")`);
