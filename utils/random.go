// Package utils provides random data generators and date helpers for the workflow.
package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// init seeds the default random source.
func init() {
	rand.Seed(time.Now().UnixNano())
}

// ---------- Name generators ----------

var firstNames = []string{
	"Mohammed", "Ahmed", "Khalid", "Omar", "Faisal",
	"Sultan", "Nasser", "Saad", "Turki", "Abdullah",
	"Fahad", "Majed", "Bandar", "Rayan", "Hamad",
	"Youssef", "Ali", "Hassan", "Ibrahim", "Mansour",
}

var lastNames = []string{
	"AlSaud", "AlQahtani", "AlGhamdi", "AlDossari", "AlShehri",
	"AlHarbi", "AlOtaibi", "AlMutairi", "AlSubaie", "AlZahrani",
	"AlMalki", "AlRashidi", "AlKhaldi", "AlTamimi", "AlEnzi",
}

// RandomName returns a random full name.
func RandomName() string {
	first := firstNames[rand.Intn(len(firstNames))]
	last := lastNames[rand.Intn(len(lastNames))]
	return fmt.Sprintf("%s %s", first, last)
}

// RandomFirstName returns a random first name only.
func RandomFirstName() string {
	return firstNames[rand.Intn(len(firstNames))]
}

// ---------- Email ----------

// RandomEmail returns a unique random email address.
func RandomEmail(domain string) string {
	if domain == "" {
		domain = "testmail.com"
	}
	ts := time.Now().UnixMilli()
	r := rand.Intn(9999)
	name := strings.ToLower(firstNames[rand.Intn(len(firstNames))])
	return fmt.Sprintf("%s+%d%d@%s", name, ts, r, domain)
}

// ---------- National ID ----------

// RandomNationalID returns a random 10-digit national ID starting with "1" or "2".
func RandomNationalID() string {
	prefix := rand.Intn(2) + 1 // 1 or 2
	suffix := rand.Intn(900000000) + 100000000
	return fmt.Sprintf("%d%d", prefix, suffix)
}

// ---------- Phone ----------

// RandomPhone returns a random Saudi mobile number (05XXXXXXXX).
func RandomPhone() string {
	return fmt.Sprintf("05%08d", rand.Intn(100000000))
}

// ---------- Company identifiers ----------

// RandomUniNumber returns a random 10-digit number starting with 700.
func RandomUniNumber() string {
	return fmt.Sprintf("700%07d", rand.Intn(10000000))
}

// RandomCRNumber returns a random 10-digit CR number starting with 700.
func RandomCRNumber() string {
	return fmt.Sprintf("700%07d", rand.Intn(10000000))
}

// ---------- Company names ----------

var companyPrefixes = []string{
	"Tamkeen", "Mashrooa", "Taqat", "Rawabi", "Nomu",
	"Binaa", "Watan", "Itqan", "Daleel", "Istithmar",
	"Takamol", "Injaz", "Masar", "Hayat", "Kaizen",
}

var companySuffixes = []string{
	"Holdings", "Ventures", "Capital", "Group", "Solutions",
	"Industries", "Trading", "Development", "Logistics", "Tech",
}

// RandomCompanyName returns a random English company name.
func RandomCompanyName() string {
	p := companyPrefixes[rand.Intn(len(companyPrefixes))]
	s := companySuffixes[rand.Intn(len(companySuffixes))]
	return fmt.Sprintf("%s %s %d", p, s, rand.Intn(999)+1)
}

// RandomCompanyNameAr returns a random Arabic company name.
func RandomCompanyNameAr() string {
	names := []string{
		"شركة التمكين", "شركة النمو", "شركة البناء",
		"شركة الإتقان", "شركة المسار", "شركة الإنجاز",
		"شركة الاستثمار", "شركة الوطن", "شركة التكامل",
	}
	return fmt.Sprintf("%s %d", names[rand.Intn(len(names))], rand.Intn(999)+1)
}

// ---------- Project names ----------

var projectPrefixes = []string{
	"Solar Farm", "Logistics Hub", "Smart City", "Health Center",
	"Tech Park", "Green Energy", "Water Treatment", "Education Hub",
	"Industrial Zone", "Residential Tower", "Mixed Use",
}

// RandomProjectName returns a random English project name.
func RandomProjectName() string {
	p := projectPrefixes[rand.Intn(len(projectPrefixes))]
	return fmt.Sprintf("%s Phase %d", p, rand.Intn(10)+1)
}

// RandomProjectNameAr returns a random Arabic project name.
func RandomProjectNameAr() string {
	names := []string{
		"مشروع الطاقة الشمسية", "مشروع المدينة الذكية", "مشروع المستشفى",
		"مشروع المجمع السكني", "مشروع المنطقة الصناعية", "مشروع البنية التحتية",
	}
	return fmt.Sprintf("%s - المرحلة %d", names[rand.Intn(len(names))], rand.Intn(10)+1)
}

// ---------- Descriptions ----------

// RandomDescription returns a random English description.
func RandomDescription() string {
	descs := []string{
		"A transformative project aimed at boosting economic growth in the region.",
		"An innovative initiative to deliver sustainable infrastructure solutions.",
		"A high-impact development project supporting Vision 2030 objectives.",
		"A strategic investment opportunity in a rapidly growing sector.",
		"A landmark project combining technology with community development.",
		"An ambitious venture targeting underserved market segments.",
	}
	return descs[rand.Intn(len(descs))]
}

// RandomDescriptionAr returns a random Arabic description.
func RandomDescriptionAr() string {
	descs := []string{
		"مشروع تحولي يهدف إلى تعزيز النمو الاقتصادي في المنطقة.",
		"مبادرة مبتكرة لتقديم حلول بنية تحتية مستدامة.",
		"مشروع تنموي عالي التأثير يدعم أهداف رؤية 2030.",
		"فرصة استثمارية استراتيجية في قطاع سريع النمو.",
	}
	return descs[rand.Intn(len(descs))]
}

// ---------- Misc ----------

// RandomURL returns a random URL.
func RandomURL(prefix string) string {
	slug := strings.ToLower(firstNames[rand.Intn(len(firstNames))])
	return fmt.Sprintf("https://%s.com/%s%d", prefix, slug, rand.Intn(9999))
}

// RandomInt returns a random integer in [min, max].
func RandomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

// RandomFloat returns a random float formatted as string with 2 decimals.
func RandomFloat(min, max float64) string {
	v := min + rand.Float64()*(max-min)
	return fmt.Sprintf("%.2f", v)
}

// RandomLatitude returns a random Saudi latitude.
func RandomLatitude() string {
	return fmt.Sprintf("%.6f", 21.0+rand.Float64()*7.0)
}

// RandomLongitude returns a random Saudi longitude.
func RandomLongitude() string {
	return fmt.Sprintf("%.6f", 39.0+rand.Float64()*12.0)
}

// RandomAddress returns a random Saudi city address.
func RandomAddress() string {
	cities := []string{"Riyadh", "Jeddah", "Dammam", "Makkah", "Madinah", "Tabuk", "Abha"}
	streets := []string{"King Fahd Rd", "Olaya St", "Tahlia St", "Prince Sultan Rd", "Al Amir Rd"}
	return fmt.Sprintf("%d %s, %s, Saudi Arabia",
		rand.Intn(9000)+1000,
		streets[rand.Intn(len(streets))],
		cities[rand.Intn(len(cities))],
	)
}

// RandomLabel generates a random label pair (AR/EN).
func RandomLabel() (ar, en string) {
	labels := []struct{ ar, en string }{
		{"وثيقة المشروع", "Project Document"},
		{"التقرير المالي", "Financial Report"},
		{"خطة العمل", "Business Plan"},
		{"دراسة الجدوى", "Feasibility Study"},
		{"التحليل البيئي", "Environmental Analysis"},
		{"العقد", "Contract"},
	}
	l := labels[rand.Intn(len(labels))]
	return l.ar, l.en
}
