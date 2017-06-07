var d3 = require('d3');
var React = require('react');

class Slice extends React.Component {
	render() {
		var center = this.props.cx + " " + this.props.cy;
		var rotateDegrees = this.props.startAngle * 180 / Math.PI;

		if (this.props.angle > Math.PI*2 - 0.00001) {
			var slice = (<circle cx={this.props.cx} cy={this.props.cy} r={this.props.radius}
				className={this.props.className}
				onClick={this.props.onClick} />);
		} else {
			var dx = Math.cos(this.props.angle)*this.props.radius - this.props.radius;
			var dy = Math.sin(this.props.angle)*this.props.radius - 0.00001;
			var large_arc_flag = this.props.angle > Math.PI ? 1 : 0;

			var slice = (<path d={"M" + center + " l " + this.props.radius + " 0 a " + this.props.radius + " " + this.props.radius + ", 0, " + large_arc_flag + ", 1, " + dx + " " + dy + " Z"}
						className={this.props.className}
						onClick={this.props.onClick} />);
		}

		return (
			<g className="chart-series" transform={"rotate(" + rotateDegrees + " " + center + ")"}>
				<title>{this.props.title}</title>
				{slice}
			</g>
		);
	}
}

class PieChart extends React.Component {
	sortedSeries(series) {
		// Return an array of the series names, from highest to lowest sums (in
		// absolute terms)

		var seriesNames = [];
		var seriesValues = {};
		for (var child in series) {
			if (series.hasOwnProperty(child)) {
				seriesNames.push(child);
				seriesValues[child] = series[child].reduce(function(accum, curr, i, arr) {
					return accum + curr;
				}, 0);
			}
		}
		seriesNames.sort(function(a, b) {
			return seriesValues[b] - seriesValues[a];
		});

		return [seriesNames, seriesValues];
	}
	render() {
		var height = 400;
		var width = 600;
		var legendWidth = 100;
		var xMargin = 70;
		var yMargin = 70;
		height -= yMargin*2;
		width -= xMargin*2;
		var radius = Math.min(height, width)/2;

		var sortedSeriesValues = this.sortedSeries(this.props.report.FlattenedSeries);
		var sortedSeries = sortedSeriesValues[0];
		var seriesValues = sortedSeriesValues[1];
		var r = d3.scaleLinear()
			.range([0, 2*Math.PI])
			.domain([0, sortedSeries.reduce(function(accum, curr, i, arr) {
				return accum + Math.abs(seriesValues[curr]);
			}, 0)]);

		var slices = [];

		// Add all the slices
		var legendMap = {};
		var childId=1;
		var startAngle = 0;
		for (var i=0; i < sortedSeries.length; i++) {
			var child = sortedSeries[i];
			var value = seriesValues[child];
			if (value == 0)
				continue;

			var sliceClasses = "chart-element chart-color" + (childId % 12);
			var self = this;
			var sliceOnClick = function() {
				var childName = child;
				var onSelectSeries = self.props.onSelectSeries;
				return function() {
					onSelectSeries(childName);
				};
			}();

			var radians = r(Math.abs(value));
			var title = child + ": " + value;

			slices.push((
				<Slice key={"slice-" + childId} title={title}
					className={sliceClasses}
					cx={width/2}
					cy={height/2}
					startAngle={startAngle}
					angle={radians}
					radius={radius}
					onClick={sliceOnClick}
				/>
			));
			legendMap[child] = childId;
			childId++;
			startAngle += radians;
		}

		var legend = [];
		for (var series in legendMap) {
			var legendClasses = "chart-color" + (legendMap[series] % 12);
			var legendY = (legendMap[series] - 1)*15;
			legend.push((
				<rect key={"legend-key-"+legendMap[series]} className={legendClasses} x={0} y={legendY} width={10} height={10}/>
			));
			legend.push((
				<text key={"legend-label-"+legendMap[series]} x={0 + 15} y={legendY + 10}>{series}</text>
			));
		}

		return (
			<svg height={height + 2*yMargin} width={width + 2*xMargin + legendWidth}>
				<g className="pie-chart" transform={"translate("+xMargin+" "+yMargin+")"}>
					{slices}
				</g>
				<g className="chart-legend" transform={"translate("+(width + 2*xMargin)+" "+yMargin+")"}>
					{legend}
				</g>
			</svg>
		);
	}
}

module.exports = PieChart;
