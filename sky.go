package main

import (
	"math"
	"time"
)

type mat3 struct {
	Yz, xz, yz float64
}

/*
function CalculateZenitalAbsolutes(turbidity, solar_zenith)
{
	var ret = {}

	var Yz = (4.0453 * turbidity - 4.9710) * Math.tan((4/9 - turbidity/120) * (Math.PI - 2*solar_zenith)) - 0.2155 * turbidity + 2.4192;
	//var Y0 = (4.0453 * turbidity - 4.9710) * Math.tan((4/9 - turbidity/120) * (Math.PI)) - 0.2155 * turbidity + 2.4192;
	Y0 = 40;
	ret.Yz = Yz / Y0;

	var z3 = Math.pow(solar_zenith, 3);
	var z2 = Math.pow(solar_zenith, 2);
	var z = solar_zenith;
	var T_vec = [turbidity*turbidity, turbidity, 1.0];

	var x = [
		 0.00166 * z3 - 0.00375 * z2 + 0.00209 * z + 0,
		-0.02903 * z3 + 0.06377 * z2 - 0.03202 * z + 0.00394,
		 0.11693 * z3 - 0.21196 * z2 + 0.06052 * z + 0.25886 ];
	ret.xz = T_vec[0]*x[0] + T_vec[1]*x[1] + T_vec[2]*x[2]; // dot(T_vec, x);

	var y = [
		 0.00275 * z3 - 0.00610 * z2 + 0.00317 * z + 0,
		-0.04214 * z3 + 0.08970 * z2 - 0.04153 * z + 0.00516,
		 0.15346 * z3 - 0.26756 * z2 + 0.06670 * z + 0.26688 ];
	ret.yz = T_vec[0]*y[0] + T_vec[1]*y[1] + T_vec[2]*y[2]; // dot(T_vec, y);

	return ret;
}
*/
func CalculateZenitalAbsolutes(turbidity, solar_zenith float64) (ret mat3) {
	Yz := (4.0453*turbidity-4.9710)*math.Tan((4/9-turbidity/120)*(math.Pi-2*solar_zenith)) - 0.2155*turbidity + 2.4192
	Y0 := 40.0
	ret.Yz = Yz / Y0

	z3 := math.Pow(solar_zenith, 3)
	z2 := math.Pow(solar_zenith, 2)
	z := solar_zenith
	T_vec := []float64{turbidity * turbidity, turbidity, 1.0}

	x := []float64{
		0.00166*z3 - 0.00375*z2 + 0.00209*z + 0,
		-0.02903*z3 + 0.06377*z2 - 0.03202*z + 0.00394,
		0.11693*z3 - 0.21196*z2 + 0.06052*z + 0.25886}
	ret.xz = T_vec[0]*x[0] + T_vec[1]*x[1] + T_vec[2]*x[2] // dot(T_vec, x);

	y := []float64{
		0.00275*z3 - 0.00610*z2 + 0.00317*z + 0,
		-0.04214*z3 + 0.08970*z2 - 0.04153*z + 0.00516,
		0.15346*z3 - 0.26756*z2 + 0.06670*z + 0.26688}
	ret.yz = T_vec[0]*y[0] + T_vec[1]*y[1] + T_vec[2]*y[2] // dot(T_vec, y);

	return
}

type Coeffs struct {
	coeffsY, coeffsx, coeffsy struct{ A, B, C, D, E float64 }
}

/*
function CalculateCoefficents(turbidity)
	{
		var ret = {};

	    var coeffsY = {};
	    coeffsY.A =  0.1787 * turbidity - 1.4630;
	    coeffsY.B = -0.3554 * turbidity + 0.4275;
	    coeffsY.C = -0.0227 * turbidity + 5.3251;
	    coeffsY.D =  0.1206 * turbidity - 2.5771;
	    coeffsY.E = -0.0670 * turbidity + 0.3703;
	    ret.coeffsY = coeffsY;

	    var coeffsx = {};
	    coeffsx.A = -0.0193 * turbidity - 0.2592;
	    coeffsx.B = -0.0665 * turbidity + 0.0008;
	    coeffsx.C = -0.0004 * turbidity + 0.2125;
	    coeffsx.D = -0.0641 * turbidity - 0.8989;
	    coeffsx.E = -0.0033 * turbidity + 0.0452;
	    ret.coeffsx = coeffsx;

	    var coeffsy = {};
	    coeffsy.A = -0.0167 * turbidity - 0.2608;
	    coeffsy.B = -0.0950 * turbidity + 0.0092;
	    coeffsy.C = -0.0079 * turbidity + 0.2102;
	    coeffsy.D = -0.0441 * turbidity - 1.6537;
	    coeffsy.E = -0.0109 * turbidity + 0.0529;
	    ret.coeffsy = coeffsy;

	    return ret;
	}
*/
func CalculateCoefficents(turbidity float64) (ret Coeffs) {
	ret.coeffsY.A = 0.1787*turbidity - 1.4630
	ret.coeffsY.B = -0.3554*turbidity + 0.4275
	ret.coeffsY.C = -0.0227*turbidity + 5.3251
	ret.coeffsY.D = 0.1206*turbidity - 2.5771
	ret.coeffsY.E = -0.0670*turbidity + 0.3703

	ret.coeffsx.A = -0.0193*turbidity - 0.2592
	ret.coeffsx.B = -0.0665*turbidity + 0.0008
	ret.coeffsx.C = -0.0004*turbidity + 0.2125
	ret.coeffsx.D = -0.0641*turbidity - 0.8989
	ret.coeffsx.E = -0.0033*turbidity + 0.0452

	ret.coeffsy.A = -0.0167*turbidity - 0.2608
	ret.coeffsy.B = -0.0950*turbidity + 0.0092
	ret.coeffsy.C = -0.0079*turbidity + 0.2102
	ret.coeffsy.D = -0.0441*turbidity - 1.6537
	ret.coeffsy.E = -0.0109*turbidity + 0.0529

	return
}

/*
	function Perez( zenith, gamma, coeffs )
	{
	    return  (1 + coeffs.A*Math.exp(coeffs.B/Math.cos(zenith))) *
	            (1 + coeffs.C*Math.exp(coeffs.D*gamma)+coeffs.E*Math.pow(Math.cos(gamma), 2));
	}
*/
func Perez(zenith, gamma float64, coeffs struct{ A, B, C, D, E float64 }) float64 {
	return (1 + coeffs.A*math.Exp(coeffs.B/math.Cos(zenith))) *
		(1 + coeffs.C*math.Exp(coeffs.D*gamma) + coeffs.E*math.Pow(math.Cos(gamma), 2))
}

/*

	function gamma_correct(v)
	{
		return Math.max(Math.min( Math.pow(v, (1/1.8)), 1), 0)
	}
*/
func gamma_correct(v float64) float64 {
	return math.Max(math.Min(math.Pow(v, (1.0/1.8)), 1.0), 0.0)
}

/*
	function Yxy_to_RGB(Y, x, y)
	{
	    var X = x/y*Y;
	    var Z = (1.0-x-y)/y*Y;
	    return {
	    	r : gamma_correct( 3.2406 * X - 1.5372 * Y - 0.4986 * Z ),
	    	g : gamma_correct(-0.9689 * X + 1.8758 * Y + 0.0415 * Z ),
	    	b : gamma_correct( 0.0557 * X - 0.2040 * Y + 1.0570 * Z )};
	}
*/

func Yxy_to_RGB(Y, x, y float64) (r, g, b float64) {
	X := x / y * Y
	Z := (1.0 - x - y) / y * Y
	r = gamma_correct(3.2406*X - 1.5372*Y - 0.4986*Z)
	g = gamma_correct(-0.9689*X + 1.8758*Y + 0.0415*Z)
	b = gamma_correct(0.0557*X - 0.2040*Y + 1.0570*Z)
	return
}

/*
	function Gamma(zenith, azimuth,   solar_zenith, solar_azimuth)
	{
	    return Math.acos(
	    	Math.sin(solar_zenith)*Math.sin(zenith)*Math.cos(azimuth-solar_azimuth)+Math.cos(solar_zenith)*Math.cos(zenith));
	}
*/
func Gamma(zenith, azimuth, solar_zenith, solar_azimuth float64) float64 {
	return math.Acos(math.Sin(solar_zenith)*math.Sin(zenith)*math.Cos(azimuth-solar_azimuth) + math.Cos(solar_zenith)*math.Cos(zenith))
}

/*
	function Calc_Sky_RGB(zenith, azimuth,   zen_abs, solar_zenith, solar_azimuth, coeffs_mtx)
	{
	    var gamma = Gamma(zenith, azimuth,  solar_zenith, solar_azimuth);
	    zenith = Math.min(zenith, Math.PI/2.0);
	    var Yp = zen_abs.Yz * Perez(zenith, gamma, coeffs_mtx.coeffsY) / Perez(0.0, solar_zenith, coeffs_mtx.coeffsY);
	    var xp = zen_abs.xz * Perez(zenith, gamma, coeffs_mtx.coeffsx) / Perez(0.0, solar_zenith, coeffs_mtx.coeffsx);
	    var yp = zen_abs.yz * Perez(zenith, gamma, coeffs_mtx.coeffsy) / Perez(0.0, solar_zenith, coeffs_mtx.coeffsy);

	    return Yxy_to_RGB(Yp, xp, yp);
	}
*/
func Calc_Sky_RGB(zenith, azimuth, solar_zenith, solar_azimuth float64, zen_abs mat3, coeffs_mtx Coeffs) (r, g, b float64) {
	gamma := Gamma(zenith, azimuth, solar_zenith, solar_azimuth)
	zenith = math.Min(zenith, math.Pi/2.0)
	Yp := zen_abs.Yz * Perez(zenith, gamma, coeffs_mtx.coeffsY) / Perez(0.0, solar_zenith, coeffs_mtx.coeffsY)
	xp := zen_abs.xz * Perez(zenith, gamma, coeffs_mtx.coeffsx) / Perez(0.0, solar_zenith, coeffs_mtx.coeffsx)
	yp := zen_abs.yz * Perez(zenith, gamma, coeffs_mtx.coeffsy) / Perez(0.0, solar_zenith, coeffs_mtx.coeffsy)

	return Yxy_to_RGB(Yp, xp, yp)

}

/*
	function scale_range(k,  k_min, k_max,  v_min, v_max)
	{
		return (k-k_min)/(k_max-k_min) * (v_max-v_min) + v_min
	}
*/
func scale_range(k, k_min, k_max, v_min, v_max float64) float64 {
	return (k-k_min)/(k_max-k_min)*(v_max-v_min) + v_min
}

/*
	function deg2rad(deg)
	{
		return deg * (Math.PI/180);
	}
*/

func deg2rad(deg float64) float64 {
	return deg * (math.Pi / 180.0)
}

func getSkyAtPoint(x, y, turbidity, latitude, longitude, time_zone_meridian float64, date time.Time, azimuthRange, zenithEnd float64) (r, g, b float64) {
	//fmt.Printf("date: %s\n", date)
	longitude = deg2rad(longitude)
	latitude = deg2rad(latitude)
	time_zone_meridian = deg2rad(time_zone_meridian)

	julian := float64(date.YearDay())
	today_running := float64(date.Hour()) + (float64(date.Minute()) / 60.0) + (float64(date.Second()) / 3600.0)
	solar_time := today_running +
		0.170*math.Sin((4*math.Pi*(julian-80))/373) -
		0.129*math.Sin((2*math.Pi*(julian-8))/355) +
		(12 * (time_zone_meridian - longitude) / math.Pi)

	declination := 0.4093 * math.Sin(2*math.Pi*(julian-81)/368)
	//fmt.Printf("solar_time: %f, julian: %f, time_zone_meridian: %f, longitude: %f, %f\n", solar_time, julian, time_zone_meridian, longitude, today_running)

	sin_l := math.Sin(latitude)
	cos_l := math.Cos(latitude)
	sin_d := math.Sin(declination)
	cos_d := math.Cos(declination)
	cos_pi_t_12 := math.Cos(math.Pi * solar_time / 12.0)
	sin_pi_t_12 := math.Sin(math.Pi * solar_time / 12.0)
	solar_zenith := math.Pi/2 - math.Asin(sin_l*sin_d-cos_l*cos_d*cos_pi_t_12)
	solar_azimuth := math.Atan2((-cos_d * sin_pi_t_12), (cos_l*sin_d - sin_l*cos_d*cos_pi_t_12))

	//fmt.Printf("sin_l: %f, cos_l: %f, sin_d: %f, cos_d: %f, cos_pi_t_12: %f, sin_pi_t_12: %f\n", sin_l, cos_l, sin_d, cos_d, cos_pi_t_12, sin_pi_t_12)
	//fmt.Printf("solar_zenith: %f, solar_azimuth: %f\n", solar_zenith, solar_azimuth)

	zen_abs := CalculateZenitalAbsolutes(turbidity, solar_zenith)
	coeffs_mtx := CalculateCoefficents(turbidity)
	//fmt.Printf("zen_abs: %+v, coeffs_mtx: %+v\n", zen_abs, coeffs_mtx)

	azimuth := deg2rad(scale_range(y, 0, 1, -azimuthRange, azimuthRange))
	zenith := deg2rad(scale_range(x, 0, 1, zenithEnd, 90))

	//fmt.Printf("solar_azimuth: %+v, solar_zenith: %+v\n", solar_azimuth, solar_zenith)

	return Calc_Sky_RGB(zenith, azimuth, solar_zenith, solar_azimuth, zen_abs, coeffs_mtx)
}
