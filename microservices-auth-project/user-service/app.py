from flask import Flask, request, jsonify
from flask_sqlalchemy import SQLAlchemy
import os
import jwt
import requests
from functools import wraps

app = Flask(__name__)

# Configuration
app.config['SQLALCHEMY_DATABASE_URI'] = os.getenv('DATABASE_URL', 'postgresql://user:password@postgres:5432/users')
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
AUTH_SERVICE_URL = os.getenv('AUTH_SERVICE_URL', 'http://auth-service:4000')

db = SQLAlchemy(app)

# Models
class UserProfile(db.Model):
    __tablename__ = 'user_profiles'
    
    id = db.Column(db.String(50), primary_key=True)  # Matches Auth Service ID
    email = db.Column(db.String(120), unique=True, nullable=False)
    full_name = db.Column(db.String(100))
    phone = db.Column(db.String(20))
    bio = db.Column(db.Text)
    created_at = db.Column(db.DateTime, server_default=db.func.now())
    updated_at = db.Column(db.DateTime, server_default=db.func.now(), onupdate=db.func.now())

    def to_dict(self):
        return {
            'id': self.id,
            'email': self.email,
            'full_name': self.full_name,
            'phone': self.phone,
            'bio': self.bio,
            'created_at': self.created_at.isoformat() if self.created_at else None
        }

# Middleware
def token_required(f):
    @wraps(f)
    def decorated(*args, **kwargs):
        token = None
        
        # Get token from header
        if 'Authorization' in request.headers:
            auth_header = request.headers['Authorization']
            if auth_header.startswith('Bearer '):
                token = auth_header.split(' ')[1]
        
        if not token:
            return jsonify({'message': 'Token is missing'}), 401
            
        try:
            # Validate token with Auth Service
            # In a real high-perf scenario, we might verify JWT signature locally
            # But calling Auth Service ensures we check the blacklist
            response = requests.post(
                f"{AUTH_SERVICE_URL}/auth/validate",
                headers={'Authorization': f"Bearer {token}"}
            )
            
            if response.status_code != 200:
                return jsonify({'message': 'Token is invalid or expired'}), 401
                
            current_user = response.json().get('user')
            
        except Exception as e:
            print(f"Auth validation error: {e}")
            return jsonify({'message': 'Authentication failed'}), 500
            
        return f(current_user, *args, **kwargs)
    
    return decorated

# Routes
@app.route('/health', methods=['GET'])
def health_check():
    return jsonify({'status': 'healthy', 'service': 'user-service'})

@app.route('/users/me', methods=['GET'])
@token_required
def get_my_profile(current_user):
    user_id = current_user.get('userId')
    
    profile = UserProfile.query.get(user_id)
    
    if not profile:
        # Create default profile if not exists (lazy creation)
        profile = UserProfile(
            id=user_id,
            email=current_user.get('email'),
            full_name=current_user.get('name', 'Unknown')
        )
        db.session.add(profile)
        db.session.commit()
        
    return jsonify(profile.to_dict())

@app.route('/users/me', methods=['PUT'])
@token_required
def update_my_profile(current_user):
    user_id = current_user.get('userId')
    data = request.get_json()
    
    profile = UserProfile.query.get(user_id)
    
    if not profile:
        return jsonify({'message': 'Profile not found'}), 404
     
    if 'full_name' in data:
        profile.full_name = data['full_name']
    if 'phone' in data:
        profile.phone = data['phone']
    if 'bio' in data:
        profile.bio = data['bio']
        
    db.session.commit()
    
    return jsonify(profile.to_dict())

# Internal endpoint for other services
@app.route('/users/<user_id>', methods=['GET'])
def get_user_by_id(user_id):
    # In production, secure this with internal API key or mTLS
    profile = UserProfile.query.get(user_id)
    if not profile:
        return jsonify({'message': 'User not found'}), 404
    return jsonify(profile.to_dict())

# Initialize DB
with app.app_context():
    db.create_all()

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
