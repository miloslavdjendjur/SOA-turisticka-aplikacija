import { Component, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import * as L from 'leaflet';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {
  private readonly API = 'http://localhost:8000';

  // --- MAPA ---
  private map: any;
  private marker: any;          // Crveni marker (Ti)
  private tourMarkers: any[] = []; // Zeleni markeri (Kljucne tacke)

  // --- STANJE ---
  currentUserId: number | null = null;
  message: string = '';
  activeTab: string = 'users';

  // --- PODACI ---
  loginId: number | null = null;
  registerData = { username: '', password: '', email: '', role: 'TOURIST', name: '', surname: '' };
  userProfile: any = null;

  tourData = { authorId: 0, name: '', description: '', difficulty: 'EASY', tags: ['ispit'], price: 0, distance: 0 };
  tours: any[] = [];

  keyPointData = { tourId: 0, name: '', description: '', latitude: 45.2671, longitude: 19.8335, image: 'url', order: 1 };
  blogData = { authorId: 0, title: '', description: '', image: '' };
  blogs: any[] = [];

  followId: number | null = null;
  followingList: number[] = [];

  simulator = { latitude: 45.2671, longitude: 19.8335 };


  cart: any = null;


  activeTourId: number | null = null;
  activeTourStatus: string = '';
  checkInterval: any = null;

  constructor(private http: HttpClient) {}

  ngOnInit() {
    this.loadTours(); 
    this.loadBlogs();
  }

  saveState() {
    if (this.currentUserId) localStorage.setItem('currentUserId', this.currentUserId.toString());
    if (this.activeTourId) localStorage.setItem('activeTourId', this.activeTourId.toString());
    else localStorage.removeItem('activeTourId');
  }

  restoreState() {

    const savedUser = localStorage.getItem('currentUserId');
    if (savedUser) {
      this.currentUserId = +savedUser;
      this.loginId = +savedUser;
      this.getProfile();
      this.loadCart(); 
    }


    const savedTour = localStorage.getItem('activeTourId');
    if (savedTour && this.currentUserId) {
      this.activeTourId = +savedTour;
      this.message = "🔄 Tura restaurirana! Nastavljamo...";
      this.setTab('tours'); 
      this.startCheckingProximity(this.activeTourId);
    }
  }


  setTab(tab: string) {
    this.activeTab = tab;
    if (tab === 'tours') {
      setTimeout(() => { 
        this.initMap(); 
        if (this.activeTourId) this.drawTourOnMap(this.activeTourId);
      }, 100);
    }
  }

  initMap() {
    if (this.map) return; 

    this.map = L.map('map').setView([45.2671, 19.8335], 13);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '© OpenStreetMap'
    }).addTo(this.map);

    this.map.on('click', (e: any) => {
      const lat = e.latlng.lat;
      const lng = e.latlng.lng;

      this.simulator.latitude = lat;
      this.simulator.longitude = lng;
      this.keyPointData.latitude = lat;
      this.keyPointData.longitude = lng;


      if (this.marker) {
        this.marker.setLatLng([lat, lng]);
      } else {
        this.marker = L.circleMarker([lat, lng], { radius: 10, color: 'red', fillColor: '#f03', fillOpacity: 0.8 }).addTo(this.map);
      }
    });
  }

  drawTourOnMap(tourId: number) {
    if (this.tours.length === 0) return;
    this.clearTourFromMap();

    const tour = this.tours.find(t => t.id === tourId);
    if (!tour || !tour.keyPoints) return;


    tour.keyPoints.forEach((kp: any) => {
      const m = L.circleMarker([kp.latitude, kp.longitude], {
        radius: 12, color: 'green', fillColor: '#0f0', fillOpacity: 0.6, weight: 2
      }).addTo(this.map).bindPopup(`<b>${kp.name}</b><br>${kp.description}`);
      this.tourMarkers.push(m);
    });

    if (tour.keyPoints.length > 0) {
      const group = L.featureGroup(this.tourMarkers);
      this.map.fitBounds(group.getBounds().pad(0.2));
    }
  }

  clearTourFromMap() {
    this.tourMarkers.forEach(m => this.map.removeLayer(m));
    this.tourMarkers = [];
  }


  login() {
    if (!this.loginId) return;
    this.currentUserId = this.loginId;
    this.message = `👋 Login ID: ${this.loginId}`;
    this.saveState();
    this.getProfile();
    this.loadCart();
  }

  register() {
    this.http.post(`${this.API}/api/users/register`, this.registerData).subscribe({
      next: (res: any) => {
        this.currentUserId = res.id;
        this.loginId = res.id;
        this.message = `✅ Registrovan ID: ${res.id}`;
        this.saveState();
        this.getProfile();
      },
      error: (err) => this.message = `❌ Greška: ${err.error || err.message}`
    });
  }

  getProfile() {
    if(!this.currentUserId) return;
    this.http.get(`${this.API}/api/users/${this.currentUserId}`).subscribe({
      next: (res) => this.userProfile = res,
      error: (err) => this.message = "⚠️ Profil nije nađen."
    });
  }

  updateProfile() {
    if(!this.currentUserId) return;
    this.http.put(`${this.API}/api/users/${this.currentUserId}`, this.userProfile).subscribe({
      next: () => this.message = "✅ Profil ažuriran!",
      error: (err) => this.message = "❌ Greška update."
    });
  }


  createBlog() {
    if(!this.currentUserId) return;
    this.blogData.authorId = this.currentUserId;
    this.http.post(`${this.API}/blogs`, this.blogData).subscribe({
      next: () => { this.message = "✅ Blog objavljen!"; this.loadBlogs(); },
      error: (err) => this.message = `❌ Greška: ${err.message}`
    });
  }

  loadBlogs() {
    this.http.get<any[]>(`${this.API}/blogs`).subscribe(data => this.blogs = data);
  }

  followUser() {
    if(!this.currentUserId || !this.followId) return;
    this.http.post(`${this.API}/followers`, { followerId: this.currentUserId, followedId: this.followId }).subscribe({
      next: () => { 
        this.message = `✅ Pratiš ID ${this.followId}`; 
        this.followingList.push(this.followId!);
      },
      error: (err) => this.message = `❌ Greška: ${err.message}`
    });
  }


  createTour() {
    if(!this.currentUserId) return;
    this.tourData.authorId = this.currentUserId;
    this.http.post(`${this.API}/tours`, this.tourData).subscribe({
      next: () => { this.message = "✅ Tura kreirana!"; this.loadTours(); },
      error: (err) => this.message = `❌ Greška: ${err.message}`
    });
  }

  addKeyPoint() {
    if(!this.keyPointData.tourId) { this.message = "⚠️ Unesi ID ture!"; return; }
    this.http.post(`${this.API}/keypoints`, this.keyPointData).subscribe({
      next: () => {
        this.message = "✅ Tačka dodata!";
        if (this.activeTourId === this.keyPointData.tourId) {
           this.loadTours(); 
           setTimeout(() => this.drawTourOnMap(this.activeTourId!), 500);
        }
      },
      error: (err) => this.message = `❌ Greška: ${err.message}`
    });
  }


  loadCart() {
    if (!this.currentUserId) return;
    this.http.get(`${this.API}/cart?touristId=${this.currentUserId}`).subscribe({
      next: (res) => this.cart = res,
      error: (err) => console.log("Korpa prazna")
    });
  }

  addToCart(tour: any) {
    if (!this.currentUserId) { this.message = "⚠️ Uloguj se!"; return; }
    const item = { touristId: this.currentUserId, tourId: tour.id, tourName: tour.name, price: tour.price };
    this.http.post(`${this.API}/cart/add`, item).subscribe({
      next: () => { this.message = `🛒 Dodato: ${tour.name}`; this.loadCart(); },
      error: (err) => this.message = `❌ Greška: ${err.message}`
    });
  }

  checkout() {
    if (!this.currentUserId) return;
    this.http.post(`${this.API}/cart/checkout`, { touristId: this.currentUserId }).subscribe({
      next: () => { this.message = "✅ Kupovina uspešna! Dobio si tokene."; this.loadCart(); },
      error: (err) => this.message = `❌ Greška: ${err.message}`
    });
  }


  startTour(tourId: number) {
    if(!this.currentUserId) return;
    this.http.post(`${this.API}/tours/start`, { touristId: this.currentUserId, tourId }).subscribe({
      next: () => {
        this.message = `🚀 Tura pokrenuta! Ciljaj zelene tačke.`;
        this.activeTourId = tourId;
        this.saveState();
        this.drawTourOnMap(tourId);
        this.startCheckingProximity(tourId);
      },
      error: (err) => this.message = `❌ Greška (Kupi turu prvo!): ${err.error.error || err.message}`
    });
  }

  endTour() {
    if(!this.activeTourId || !this.currentUserId) return;
    this.http.post(`${this.API}/tours/end`, { touristId: this.currentUserId, tourId: this.activeTourId }).subscribe({
      next: () => {
        this.message = "🏁 Tura završena!";
        this.activeTourId = null;
        this.activeTourStatus = "";
        this.saveState();
        this.clearTourFromMap();
        if (this.checkInterval) clearInterval(this.checkInterval);
      },
      error: (err) => this.message = `❌ Greška: ${err.message}`
    });
  }

  updatePosition() {
    if(!this.currentUserId) return;
    const body = { touristId: this.currentUserId, ...this.simulator };
    this.http.post(`${this.API}/position`, body).subscribe({
      next: () => { if (!this.activeTourId) this.message = "📍 Pozicija ažurirana!"; },
      error: (err) => this.message = `❌ Greška: ${err.message}`
    });
  }

  startCheckingProximity(tourId: number) {
    if (this.checkInterval) clearInterval(this.checkInterval);
    this.checkInterval = setInterval(() => {
      if (!this.activeTourId) return;
      this.updatePosition(); 
      this.http.post(`${this.API}/tours/check`, { touristId: this.currentUserId, tourId }).subscribe({
        next: (res: any) => {
          if (res.nearKeyPoint) this.activeTourStatus = `📍 POGODAK! Stigao si do: ${res.pointName}`;
          else this.activeTourStatus = "🚶 Hodaj ka ZELENOJ tački...";
        }
      });
    }, 5000);
  }

  loadTours() {
    this.http.get<any[]>(`${this.API}/tours`).subscribe(data => {
      this.tours = data;
      this.restoreState(); 
    });
  }
}