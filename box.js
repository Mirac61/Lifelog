<select name="type" hx-get="/heatmap" hx-target="#chart" hx-trigger="change">
  <option value="commit">Commits</option>
</select>

<div id="chart">
  <!-- hier landet das SVG -->
</div>
