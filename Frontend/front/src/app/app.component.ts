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
  private marker: any;
  private tourMarkers: any[] = [];

  // --- STANJE ---
  currentUserId: number | null = null;
  message: string = '';
  activeTab: string = 'auth';

  // --- PODACI ---
  loginData = { username: '', password: '' };
  registerData = { 
    username: '', 
    password: '', 
    email: '', 
    role: 'TOURIST', 
    name: '', 
    surname: '' 
  };
  
  userProfile: any = null;

  // Ture
  tourData = { authorId: 0, name: '', description: '', difficulty: 'EASY', tags: ['ispit'], price: 0, distance: 0 };
  tours: any[] = [];

  // Ključne tačke
  keyPointData = { tourId: 0, name: '', description: '', latitude: 45.2671, longitude: 19.8335, image: 'url', order: 1 };
  
  // Blog & Komentari
  blogData = { authorId: 0, title: '', description: '', image: '' };
  blogs: any[] = [];
  newCommentText: string = ''; 

  // Follow & Preporuke
  followId: number | null = null;
  followingList: number[] = [];
  recommendations: number[] = []; 

  // Simulator
  simulator = { latitude: 45.2671, longitude: 19.8335 };
  cart: any = null;

  // Aktivna Tura (Tačka 17)
  activeTourId: number | null = null;
  activeTourStatus: string = '';
  checkInterval: any = null;

  constructor(private http: HttpClient) {}

  ngOnInit() {
    this.loadTours(); 
  }

  // --- PERSISTENCE ---
  saveState() {
    if (this.currentUserId) localStorage.setItem('currentUserId', this.currentUserId.toString());
    if (this.activeTourId) localStorage.setItem('activeTourId', this.activeTourId.toString());
    else localStorage.removeItem('activeTourId');
  }

  restoreState() {
    const savedUser = localStorage.getItem('currentUserId');
    if (savedUser) {
      this.currentUserId = +savedUser;
      this.getProfile();
      this.loadCart();
      this.loadFollowingList(); 
      this.loadBlogs();
    }

    const savedTour = localStorage.getItem('activeTourId');
    if (savedTour && this.currentUserId) {
      this.activeTourId = +savedTour;
      this.setTab('tours'); 
      this.startCheckingProximity(this.activeTourId);
    }
  }

  // --- NAVIGACIJA I MAPA ---
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
      this.simulator.latitude = e.latlng.lat;
      this.simulator.longitude = e.latlng.lng;
      this.keyPointData.latitude = e.latlng.lat;
      this.keyPointData.longitude = e.latlng.lng;
      if (this.marker) this.marker.setLatLng(e.latlng);
      else this.marker = L.circleMarker(e.latlng, { radius: 10, color: 'red', fillColor: '#f03', fillOpacity: 0.8 }).addTo(this.map);
    });
  }

  drawTourOnMap(tourId: number) {
    if (this.tours.length === 0) return;
    this.clearTourFromMap();
    const tour = this.tours.find(t => t.id === tourId);
    if (!tour || !tour.keyPoints) return;
    tour.keyPoints.forEach((kp: any) => {
      const m = L.circleMarker([kp.latitude, kp.longitude], { radius: 12, color: 'green', fillColor: '#0f0', fillOpacity: 0.6, weight: 2 }).addTo(this.map).bindPopup(`<b>${kp.name}</b>`);
      this.tourMarkers.push(m);
    });
    if (tour.keyPoints.length > 0) this.map.fitBounds(L.featureGroup(this.tourMarkers).getBounds().pad(0.2));
  }

  clearTourFromMap() {
    this.tourMarkers.forEach(m => this.map.removeLayer(m));
    this.tourMarkers = [];
  }

  // --- AUTH ---
  login() {
    this.http.post(`${this.API}/api/users/login`, this.loginData).subscribe({
      next: (res: any) => {
        this.currentUserId = res.id;
        this.userProfile = res;
        this.saveState();
        this.loadCart();
        this.loadFollowingList();
        this.loadBlogs();
        this.setTab('profile');
      },
      error: () => this.message = "❌ Pogrešan login!"
    });
  }

  logout() {
    this.currentUserId = null;
    this.userProfile = null;
    this.activeTourId = null;
    this.cart = null;
    if (this.checkInterval) clearInterval(this.checkInterval);
    localStorage.clear();
    this.activeTab = 'auth';
  }

  register() {
    this.http.post(`${this.API}/api/users/register`, this.registerData).subscribe({
      next: () => this.message = "✅ Registracija uspešna!",
      error: (err) => this.message = `❌ Greška: ${err.message}`
    });
  }

  getProfile() {
    if(!this.currentUserId) return;
    this.http.get(`${this.API}/api/users/${this.currentUserId}`).subscribe(res => this.userProfile = res);
  }

  updateProfile() {
    if(!this.currentUserId) return;
    this.http.put(`${this.API}/api/users/${this.currentUserId}`, this.userProfile).subscribe(() => this.message = "✅ Profil sačuvan!");
  }

  onFileSelected(event: any) {
    const file: File = event.target.files[0];
    if (file) {
      const reader = new FileReader();
      reader.readAsDataURL(file);
      reader.onload = () => { this.userProfile.profilePicture = reader.result as string; };
    }
  }

  // --- SOCIAL ---
  createBlog() {
    if(!this.currentUserId) return;
    this.blogData.authorId = this.currentUserId;
    this.http.post(`${this.API}/blogs`, this.blogData).subscribe(() => { this.message = "✅ Blog objavljen!"; this.loadBlogs(); });
  }

  loadBlogs() {
    if(!this.currentUserId) return;
    this.http.get<any[]>(`${this.API}/blogs?userId=${this.currentUserId}`).subscribe(data => {
      this.blogs = data;
      this.blogs.forEach(b => this.loadComments(b));
    });
  }

  loadComments(blog: any) {
    this.http.get<any[]>(`${this.API}/comments?blogId=${blog.id}`).subscribe(res => blog.comments = res);
  }

  addComment(blogId: number) {
    if (!this.currentUserId) return;
    const comment = { blogId: blogId, authorId: this.currentUserId, text: this.newCommentText };
    this.http.post(`${this.API}/comments`, comment).subscribe({
      next: () => { this.message = "✅ Komentar dodat!"; this.loadBlogs(); this.newCommentText = ''; },
      error: (err) => {
         if (err.status === 403) this.message = "⛔ Moraš pratiti autora!";
         else this.message = "❌ Greška.";
      }
    });
  }

  followUser(targetId: number | null = null) {
    const idToFollow = targetId ? targetId : this.followId;
    if(!this.currentUserId || !idToFollow) return;
    this.http.post(`${this.API}/followers`, { followerId: this.currentUserId, followedId: idToFollow }).subscribe({
      next: () => { 
        this.message = `✅ Pratiš ID ${idToFollow}`; 
        this.followingList.push(idToFollow);
        this.loadRecommendations(); 
        this.loadBlogs();
      },
      error: (err) => this.message = `❌ Greška: ${err.message}`
    });
  }

  loadFollowingList() {
    if(!this.currentUserId) return;
    this.http.get<number[]>(`${this.API}/followers?userId=${this.currentUserId}`).subscribe(ids => {
        this.followingList = ids || [];
        this.loadRecommendations();
    });
  }

  loadRecommendations() {
    if(!this.currentUserId) return;
    this.http.get<number[]>(`${this.API}/followers/recommendations?userId=${this.currentUserId}`)
      .subscribe(data => this.recommendations = data);
  }

  // --- TOUR EXECUTION (Tačka 17) ---
  startTour(tourId: number) {
    if(!this.currentUserId) return;
    // Preduslov: Provera kupovine se vrši na backendu unutar /tours/start endpointa
    this.http.post(`${this.API}/tours/start`, { touristId: this.currentUserId, tourId }).subscribe({
      next: () => {
        this.activeTourId = tourId;
        this.saveState();
        this.drawTourOnMap(tourId);
        this.startCheckingProximity(tourId); // Pokretanje intervala provere lokacije
        this.message = "🚀 Tura započeta!";
      },
      error: (err) => this.message = `❌ ${err.error.error || 'Moraš kupiti turu!'}`
    });
  }

  endTour() {
    if(!this.activeTourId || !this.currentUserId) return;
    this.http.post(`${this.API}/tours/end`, { touristId: this.currentUserId, tourId: this.activeTourId, status: "COMPLETED" }).subscribe(() => {
      this.activeTourId = null;
      this.activeTourStatus = "COMPLETED";
      this.saveState();
      this.clearTourFromMap();
      if (this.checkInterval) clearInterval(this.checkInterval); 
      this.message = "🏁 Tura završena/napuštena!";
    });
  }

  abandonTour() {
    if(!this.activeTourId || !this.currentUserId) return;
    this.http.post(`${this.API}/tours/end`, { touristId: this.currentUserId, tourId: this.activeTourId, status: "ABANDONED" }).subscribe(() => {
      this.activeTourId = null;
      this.activeTourStatus = "ABANDONED";
      this.saveState();
      this.clearTourFromMap();
      if (this.checkInterval) clearInterval(this.checkInterval); 
      this.message = "🏁 Tura završena/napuštena!";
    });
  }


  updatePosition() {
    if(!this.currentUserId) return;
    this.http.post(`${this.API}/position`, { touristId: this.currentUserId, ...this.simulator }).subscribe();
  }

  startCheckingProximity(tourId: number) {
    if (this.checkInterval) clearInterval(this.checkInterval);
    
    this.checkInterval = setInterval(() => {
      this.updatePosition(); 
      

      this.http.post(`${this.API}/tours/check`, { touristId: this.currentUserId, tourId }).subscribe((res: any) => {
        if (res.nearKeyPoint) {
          this.activeTourStatus = `📍 STIGAO SI NA: ${res.pointName}`;
 
        } else {
          this.activeTourStatus = "🚶 Kreći se ka sledećoj tački...";
        }
      });
    }, 10000); // Interval od 10 sekundi
  }

  createTour() {
    if(!this.currentUserId) return;
    this.tourData.authorId = this.currentUserId;
    this.http.post(`${this.API}/tours`, this.tourData).subscribe(() => { this.message = "✅ Tura kreirana!"; this.loadTours(); });
  }

  addKeyPoint() {
    this.http.post(`${this.API}/keypoints`, this.keyPointData).subscribe(() => {
        this.message = "✅ Tačka dodata!";
        this.loadTours();
    });
  }

  loadTours() {
    this.http.get<any[]>(`${this.API}/tours`).subscribe(data => { this.tours = data; this.restoreState(); });
  }

  loadCart() {
    if (!this.currentUserId) return;
    this.http.get(`${this.API}/cart?touristId=${this.currentUserId}`).subscribe((res) => this.cart = res);
  }

  addToCart(tour: any) {
    if (!this.currentUserId) return;
    this.http.post(`${this.API}/cart/add`, { touristId: this.currentUserId, tourId: tour.id, tourName: tour.name, price: tour.price })
      .subscribe(() => { this.message = `🛒 Dodato!`; this.loadCart(); });
  }

  checkout() {
    if (!this.currentUserId) return;
    this.http.post(`${this.API}/cart/checkout`, { touristId: this.currentUserId }).subscribe(() => { this.message = "✅ Kupljeno!"; this.loadCart(); });
  }
}