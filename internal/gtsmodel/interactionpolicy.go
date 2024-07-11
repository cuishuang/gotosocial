// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package gtsmodel

// A policy URI is GoToSocial's internal representation of
// one ActivityPub URI for an Actor or a Collection of Actors,
// specific to the domain of enforcing interaction policies.
//
// A PolicyValue can be stored in the database either as one
// of the Value constants defined below (to save space), OR as
// a full-fledged ActivityPub URI.
//
// A PolicyValue should be translated to the canonical string
// value of the represented URI when federating an item, or
// from the canonical string value of the URI when receiving
// or retrieving an item.
//
// For example, if the PolicyValue `followers` was being
// federated outwards in an interaction policy attached to an
// item created by the actor `https://example.org/users/someone`,
// then it should be translated to their followers URI when sent,
// eg., `https://example.org/users/someone/followers`.
//
// Likewise, if GoToSocial receives an item with an interaction
// policy containing `https://example.org/users/someone/followers`,
// and the item was created by `https://example.org/users/someone`,
// then the followers URI would be converted to `followers`
// for internal storage.
type PolicyValue string

const (
	// Stand-in for ActivityPub magic public URI,
	// which encompasses every possible Actor URI.
	PolicyValuePublic PolicyValue = "public"
	// Stand-in for the Followers Collection of
	// the item owner's Actor.
	PolicyValueFollowers PolicyValue = "followers"
	// Stand-in for the Following Collection of
	// the item owner's Actor.
	PolicyValueFollowing PolicyValue = "following"
	// Stand-in for the Mutuals Collection of
	// the item owner's Actor.
	//
	// (TODO: Reserved, currently unused).
	PolicyValueMutuals PolicyValue = "mutuals"
	// Stand-in for Actor URIs tagged in the item.
	PolicyValueMentioned PolicyValue = "mentioned"
	// Stand-in for the Actor URI of the item owner.
	PolicyValueAuthor PolicyValue = "author"
)

// FeasibleForVisibility returns true if the PolicyValue could feasibly
// be set in a policy for an item with the given visibility, otherwise
// returns false.
//
// For example, PolicyValuePublic could not be set in a policy for an
// item with visibility FollowersOnly, but could be set in a policy
// for an item with visibility Public or Unlocked.
//
// This is not prescriptive, and should be used only to guide policy
// choices. Eg., if a remote instance wants to do something wacky like
// set "anyone can interact with this status" for a Direct visibility
// status, that's their business; our normal visibility filtering will
// prevent users on our instance from actually being able to interact
// unless they can see the status anyway.
func (p PolicyValue) FeasibleForVisibility(v Visibility) bool {
	switch p {

	// Mentioned and self Values are
	// feasible for any visibility.
	case PolicyValueAuthor,
		PolicyValueMentioned:
		return true

	// Followers/following/mutual Values
	// are only feasible for items with
	// followers visibility and higher.
	case PolicyValueFollowers,
		PolicyValueFollowing:
		return v == VisibilityFollowersOnly ||
			v == VisibilityPublic ||
			v == VisibilityUnlocked

	// Public policy Value only feasible
	// for items that are To or CC public.
	case PolicyValuePublic:
		return v == VisibilityUnlocked ||
			v == VisibilityPublic

	// Any other combo
	// is probably fine.
	default:
		return true
	}
}

type PolicyValues []PolicyValue

// PolicyResult represents the result of
// checking an Actor URI and interaction
// type against the conditions of an
// InteractionPolicy to determine if that
// interaction is permitted.
type PolicyResult int

const (
	// Interaction is forbidden for this
	// PolicyValue + interaction combination.
	PolicyResultForbidden PolicyResult = iota
	// Interaction is conditionally permitted
	// for this PolicyValue + interaction combo,
	// pending approval by the item owner.
	PolicyResultWithApproval
	// Interaction is permitted for this
	// PolicyValue + interaction combination.
	PolicyResultPermitted
)

// An InteractionPolicy determines which
// interactions will be accepted for an
// item, and according to what rules.
type InteractionPolicy struct {
	// Conditions in which a Like
	// interaction will be accepted
	// for an item with this policy.
	CanLike PolicyRules
	// Conditions in which a Reply
	// interaction will be accepted
	// for an item with this policy.
	CanReply PolicyRules
	// Conditions in which an Announce
	// interaction will be accepted
	// for an item with this policy.
	CanAnnounce PolicyRules
}

// PolicyRules represents the rules according
// to which a certain interaction is permitted
// to various Actor and Actor Collection URIs.
type PolicyRules struct {
	// Always is for PolicyValues who are
	// permitted to do an interaction
	// without requiring approval.
	Always PolicyValues
	// WithApproval is for PolicyValues who
	// are conditionally permitted to do
	// an interaction, pending approval.
	WithApproval PolicyValues
}

// Returns the default interaction policy
// for the given visibility level.
func DefaultInteractionPolicyFor(v Visibility) *InteractionPolicy {
	switch v {
	case VisibilityPublic:
		return DefaultInteractionPolicyPublic()
	case VisibilityUnlocked:
		return DefaultInteractionPolicyUnlocked()
	case VisibilityFollowersOnly, VisibilityMutualsOnly:
		return DefaultInteractionPolicyFollowersOnly()
	case VisibilityDirect:
		return DefaultInteractionPolicyDirect()
	default:
		panic("visibility " + v + " not recognized")
	}
}

// Returns the default interaction policy
// for a post with visibility of public.
func DefaultInteractionPolicyPublic() *InteractionPolicy {
	// Anyone can like.
	canLikeAlways := make(PolicyValues, 1)
	canLikeAlways[0] = PolicyValuePublic

	// Unused, set empty.
	canLikeWithApproval := make(PolicyValues, 0)

	// Anyone can reply.
	canReplyAlways := make(PolicyValues, 1)
	canReplyAlways[0] = PolicyValuePublic

	// Unused, set empty.
	canReplyWithApproval := make(PolicyValues, 0)

	// Anyone can announce.
	canAnnounceAlways := make(PolicyValues, 1)
	canAnnounceAlways[0] = PolicyValuePublic

	// Unused, set empty.
	canAnnounceWithApproval := make(PolicyValues, 0)

	return &InteractionPolicy{
		CanLike: PolicyRules{
			Always:       canLikeAlways,
			WithApproval: canLikeWithApproval,
		},
		CanReply: PolicyRules{
			Always:       canReplyAlways,
			WithApproval: canReplyWithApproval,
		},
		CanAnnounce: PolicyRules{
			Always:       canAnnounceAlways,
			WithApproval: canAnnounceWithApproval,
		},
	}
}

// Returns the default interaction policy
// for a post with visibility of unlocked.
func DefaultInteractionPolicyUnlocked() *InteractionPolicy {
	// Same as public (for now).
	return DefaultInteractionPolicyPublic()
}

// Returns the default interaction policy for
// a post with visibility of followers only.
func DefaultInteractionPolicyFollowersOnly() *InteractionPolicy {
	// Self, followers and mentioned can like.
	canLikeAlways := make(PolicyValues, 3)
	canLikeAlways[0] = PolicyValueAuthor
	canLikeAlways[1] = PolicyValueFollowers
	canLikeAlways[2] = PolicyValueMentioned

	// Unused, set empty.
	canLikeWithApproval := make(PolicyValues, 0)

	// Self, followers and mentioned can reply.
	canReplyAlways := make(PolicyValues, 3)
	canReplyAlways[0] = PolicyValueAuthor
	canReplyAlways[1] = PolicyValueFollowers
	canReplyAlways[2] = PolicyValueMentioned

	// Unused, set empty.
	canReplyWithApproval := make(PolicyValues, 0)

	// Only self can announce.
	canAnnounceAlways := make(PolicyValues, 1)
	canAnnounceAlways[0] = PolicyValueAuthor

	// Unused, set empty.
	canAnnounceWithApproval := make(PolicyValues, 0)

	return &InteractionPolicy{
		CanLike: PolicyRules{
			Always:       canLikeAlways,
			WithApproval: canLikeWithApproval,
		},
		CanReply: PolicyRules{
			Always:       canReplyAlways,
			WithApproval: canReplyWithApproval,
		},
		CanAnnounce: PolicyRules{
			Always:       canAnnounceAlways,
			WithApproval: canAnnounceWithApproval,
		},
	}
}

// Returns the default interaction policy
// for a post with visibility of direct.
func DefaultInteractionPolicyDirect() *InteractionPolicy {
	// Mentioned and self can always like.
	canLikeAlways := make(PolicyValues, 2)
	canLikeAlways[0] = PolicyValueAuthor
	canLikeAlways[1] = PolicyValueMentioned

	// Unused, set empty.
	canLikeWithApproval := make(PolicyValues, 0)

	// Mentioned and self can always reply.
	canReplyAlways := make(PolicyValues, 2)
	canReplyAlways[0] = PolicyValueAuthor
	canReplyAlways[1] = PolicyValueMentioned

	// Unused, set empty.
	canReplyWithApproval := make(PolicyValues, 0)

	// Only self can announce.
	canAnnounceAlways := make(PolicyValues, 1)
	canAnnounceAlways[0] = PolicyValueAuthor

	// Unused, set empty.
	canAnnounceWithApproval := make(PolicyValues, 0)

	return &InteractionPolicy{
		CanLike: PolicyRules{
			Always:       canLikeAlways,
			WithApproval: canLikeWithApproval,
		},
		CanReply: PolicyRules{
			Always:       canReplyAlways,
			WithApproval: canReplyWithApproval,
		},
		CanAnnounce: PolicyRules{
			Always:       canAnnounceAlways,
			WithApproval: canAnnounceWithApproval,
		},
	}
}
